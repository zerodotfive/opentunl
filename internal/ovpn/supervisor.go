package ovpn

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"slices"
	"syscall"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tooolbox/logstreamer"
)

type Supervisor struct {
	nodeName     string
	configDir    string
	configPrefix string
	postgres     *pgxpool.Pool
	servers      map[uuid.UUID]*exec.Cmd
}

func NewSupervisor(configDir string, configPrefix string, postgres *pgxpool.Pool) *Supervisor {
	nodeName, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	return &Supervisor{
		nodeName:     nodeName,
		configDir:    configDir,
		configPrefix: configPrefix,
		postgres:     postgres,
	}
}

func (s *Supervisor) getConfigPath(serverID uuid.UUID) string {
	return path.Join(s.configDir, s.configPrefix+"-"+serverID.String()+".conf")
}

func (s *Supervisor) getDesiredServersIDs(ctx context.Context) ([]uuid.UUID, error) {
	var desired []uuid.UUID
	serverProvider := NewServerProvider(ctx, s.postgres)

	servers, err := serverProvider.ListServerConfigs()
	if err != nil {
		return nil, err
	}
	for _, svr := range servers {
		state, err := s.getState(ctx, svr.ID)
		if err != nil {
			return nil, err
		}

		if state.Enabled {
			desired = append(desired, svr.ID)
		}
	}

	return desired, nil
}

func (s *Supervisor) getUnwantedServersIDs(desired []uuid.UUID) []uuid.UUID {
	unwanted := make([]uuid.UUID, 0, len(s.servers))

	for serverID, _ := range s.servers {
		if !slices.Contains(desired, serverID) {
			unwanted = append(unwanted, serverID)
		}
	}

	return unwanted
}

func (s *Supervisor) writeServerConfigFile(ctx context.Context, serverID uuid.UUID) error {
	serverConfigFilename := s.getConfigPath(serverID)

	serverProvider := NewServerProvider(ctx, s.postgres)

	server, err := serverProvider.GetServerConfig(serverID)
	if err != nil {
		return err
	}

	if err := server.WriteServerConfigFile(serverConfigFilename); err != nil {
		return err
	}

	return nil
}

func (s *Supervisor) getCmd(ctx context.Context, serverID uuid.UUID) (*exec.Cmd, error) {
	serverConfigFilename := s.getConfigPath(serverID)

	err := s.writeServerConfigFile(ctx, serverID)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("openvpn", "--config", serverConfigFilename)

	logger := log.New(os.Stdout, "[openvpn "+serverID.String()+"] ", log.LstdFlags)
	cmd.Stdout = logstreamer.NewLogstreamer(logger, "", false)
	cmd.Stderr = logstreamer.NewLogstreamer(logger, "", false)

	return cmd, nil
}

func (s *Supervisor) getState(ctx context.Context, serverID uuid.UUID) (*State, error) {
	serverProvider := NewServerProvider(ctx, s.postgres)

	server, err := serverProvider.GetServerConfig(serverID)
	if err != nil {
		return nil, err
	}

	return server.GetState(s.nodeName)
}

func (s *Supervisor) heartbeat(ctx context.Context, serverID uuid.UUID, lastPID int) error {
	state, err := s.getState(ctx, serverID)
	if err != nil {
		return err
	}

	return state.heartbeat(ctx, s.nodeName, lastPID)
}

func (s *Supervisor) checkReloadScheduled(ctx context.Context, serverID uuid.UUID) (bool, error) {
	state, err := s.getState(ctx, serverID)
	if err != nil {
		return false, err
	}

	return state.ReloadScheduled, nil
}

func (s *Supervisor) unscheduleReload(ctx context.Context, serverID uuid.UUID) error {
	state, err := s.getState(ctx, serverID)
	if err != nil {
		return err
	}

	return state.unscheduleReload(ctx, s.nodeName)
}

func (s *Supervisor) solveReload(ctx context.Context, cmd exec.Cmd, serverID uuid.UUID) {
	reloadScheduled, err := s.checkReloadScheduled(ctx, serverID)
	if err != nil {
		log.Printf("error checking reload scheduled for server %s: %s", serverID, err.Error())
	}

	if reloadScheduled {
		err := s.writeServerConfigFile(ctx, serverID)
		if err != nil {
			log.Printf("error writing config file for server %s: %s", serverID, err.Error())
		}

		if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
			log.Printf("error sending reload scheduled for server %s: %s", serverID, err.Error())
		}
	}

	if err := s.unscheduleReload(ctx, serverID); err != nil {
		log.Printf("error unscheduling reload for server %s: %s", serverID, err.Error())
	}
}

func (s *Supervisor) reconcile(ctx context.Context, unwanted []uuid.UUID, desired []uuid.UUID) {
	newList := make(map[uuid.UUID]*exec.Cmd, len(desired))

	for _, serverID := range unwanted {
		cmd, ok := s.servers[serverID]
		if !ok {
			continue
		}

		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Kill)
		}

		_ = os.RemoveAll(s.getConfigPath(serverID))
	}

	for _, serverID := range desired {
		cmd, ok := s.servers[serverID]
		if ok {
			if cmd.Process != nil && cmd.ProcessState == nil {
				if err := s.heartbeat(ctx, serverID, cmd.Process.Pid); err != nil {
					log.Printf("error sending heartbeat for server %s: %s", serverID, err.Error())
				}

				s.solveReload(ctx, *cmd, serverID)

				newList[serverID] = cmd
				continue
			}
		}

		c, err := s.getCmd(ctx, serverID)
		if err != nil {
			log.Printf("error reconciling server %s: %s", serverID, err.Error())
		}

		cmd = c

		if cmd == nil {
			continue
		}
		if err := cmd.Start(); err != nil {
			log.Printf("error reconciling server %s: %s", serverID, err.Error())
		}

		go func() { _ = cmd.Wait() }()

		newList[serverID] = cmd
	}

	s.servers = newList
}

func (s *Supervisor) loop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	desiredServers, err := s.getDesiredServersIDs(ctx)
	if err != nil {
		return err
	}

	unwantedServers := s.getUnwantedServersIDs(desiredServers)

	s.reconcile(ctx, unwantedServers, desiredServers)

	return nil
}

func (s *Supervisor) Start() {
	go func() {
		if err := s.loop(); err != nil {
			log.Printf("Error reloading servers: %s", err.Error())
		}

		for _ = range time.NewTicker(30 * time.Second).C {
			if err := s.loop(); err != nil {
				log.Printf("Error reloading servers: %s", err.Error())
			}
		}
	}()
}
