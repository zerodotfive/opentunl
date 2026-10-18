package ovpn

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type State struct {
	serverConfig    *ServerConfig
	ServerID        uuid.UUID `db:"server_id" json:"server_id"`
	NodeName        string    `db:"node_name" json:"node_name"`
	Running         bool      `json:"running"`
	LastSeen        time.Time `db:"last_seen"`
	LastPid         int       `db:"last_pid" json:"last_pid"`
	Enabled         bool      `db:"enabled" json:"enabled"`
	ReloadScheduled bool      `db:"reload_scheduled" json:"reload_scheduled"`
}

func (c *ServerConfig) ListStates() ([]State, error) {
	var states []State

	rows, err := c.provider.postgres.Query(c.provider.ctx, `
			SELECT
				n.name AS node_name,
				COALESCE(ss.last_seen, 'epoch'::timestamptz) AS last_seen,
				COALESCE(ss.last_pid, 0) AS last_pid,
				(ss.server_id IS NOT NULL) AS enabled,			
				COALESCE(ss.reload_scheduled, false) AS reload_scheduled
			FROM nodes n
			LEFT JOIN server_states ss
				ON ss.node_name = n.name
				AND ss.server_id = $1;
		`, c.ID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		s := State{
			serverConfig: c,
			ServerID:     c.ID,
		}

		err := rows.Scan(&s.NodeName, &s.LastSeen, &s.LastPid, &s.Enabled, &s.ReloadScheduled)
		if err != nil {
			return nil, err
		}

		s.Running = s.LastSeen.After(time.Now().Add(-time.Minute)) && s.LastPid != 0

		states = append(states, s)
	}

	return states, nil
}

func (c *ServerConfig) GetState(nodeName string) (*State, error) {
	s := State{
		serverConfig:    c,
		NodeName:        nodeName,
		ServerID:        c.ID,
		Enabled:         false,
		ReloadScheduled: false,
	}

	row := c.provider.postgres.QueryRow(c.provider.ctx, `
			SELECT
				COALESCE(last_seen, 'epoch'::timestamptz) AS last_seen,
				COALESCE(last_pid, 0) AS last_pid,
				COALESCE(reload_scheduled, false) AS reload_scheduled
			    FROM server_states WHERE node_name=$1 AND server_id=$2;
		`,
		nodeName, c.ID,
	)

	err := row.Scan(&s.LastSeen, &s.LastPid, &s.ReloadScheduled)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if err == nil {
		s.Enabled = true
		s.Running = s.LastSeen.After(time.Now().Add(-time.Minute)) && s.LastPid != 0
	}

	return &s, nil
}

func (s *State) Enable() error {
	_, err := s.serverConfig.provider.postgres.Exec(s.serverConfig.provider.ctx, `
		INSERT INTO server_states (node_name, server_id) VALUES ($1, $2);`,
		s.NodeName, s.ServerID,
	)

	return err
}

func (s *State) Disable() error {
	_, err := s.serverConfig.provider.postgres.Exec(s.serverConfig.provider.ctx, `
		DELETE FROM server_states WHERE node_name=$1 AND server_id=$2;`,
		s.NodeName, s.ServerID,
	)

	return err
}

func (s *State) heartbeat(ctx context.Context, nodeName string, lastPID int) error {
	_, err := s.serverConfig.provider.postgres.Exec(ctx, `
		UPDATE server_states SET last_seen=now(), last_pid=$3 WHERE node_name=$1 AND server_id=$2;`,
		nodeName, s.ServerID, lastPID,
	)

	return err
}

func (s *State) unscheduleReload(ctx context.Context, nodeName string) error {
	_, err := s.serverConfig.provider.postgres.Exec(ctx, `
		UPDATE server_states SET reload_scheduled=false WHERE node_name=$1 AND server_id=$2;`,
		nodeName, s.ServerID,
	)

	return err
}
