package ovpn

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"slices"
	"text/template"
	"uuid"

	"github.com/zerodotfive/opentunl/internal/ca"
	"github.com/zerodotfive/opentunl/internal/dh"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerProvider struct {
	ctx      context.Context
	postgres *pgxpool.Pool
}

type ServerConfigShort struct {
	provider            *ServerProvider
	ID                  uuid.UUID `db:"id" json:"id"`
	Name                string    `db:"name" json:"name"`
	RootCertificateID   uuid.UUID `db:"root_certificate_id" json:"root_certificate_id"`
	ServerCertificateID uuid.UUID `db:"server_certificate_id" json:"server_certificate_id"`
	DHID                uuid.UUID `db:"dh_id" json:"dh_id"`
	GroupsAllowed       []string  `db:"groups_allowed" json:"groups_allowed"`
}
type ServerConfig struct {
	ServerConfigShort
	ServerConfigTemplate string `db:"server_config_template" json:"server_config_template"`
	ClientConfigTemplate string `db:"client_config_template" json:"client_config_template"`
}

func NewServerProvider(ctx context.Context, postgres *pgxpool.Pool) *ServerProvider {
	return &ServerProvider{
		ctx:      ctx,
		postgres: postgres,
	}
}

func (s *ServerProvider) NewServer(name string, serverCertificateID uuid.UUID, dhID uuid.UUID, serverConfigTemplate string, clientConfigTemplate string, groupsAllowed []string) (*ServerConfig, error) {
	id := uuid.New()

	p := ca.NewProvider(s.ctx, s.postgres)

	rootCert, err := p.GetRootByServerCertificateID(serverCertificateID)
	if err != nil {
		return nil, err
	}

	serverConfig := ServerConfig{
		provider:             s,
		ID:                   id,
		Name:                 name,
		RootCertificateID:    rootCert.ID,
		ServerCertificateID:  serverCertificateID,
		DHID:                 dhID,
		ServerConfigTemplate: serverConfigTemplate,
		ClientConfigTemplate: clientConfigTemplate,
		GroupsAllowed:        groupsAllowed,
	}
	gu, _ := json.Marshal(groupsAllowed)

	log.Printf("creating server: %s", gu)

	_, err = s.postgres.Exec(s.ctx, `
		INSERT INTO server_configs
			(id, name, root_certificate_id, server_certificate_id, dh_id, server_config_template, client_config_template, groups_allowed)
			VALUES
			(@id, @name, @root_certificate_id, @server_certificate_id, @dh_id, @server_config_template, @client_config_template, @groups_allowed)`,
		pgx.NamedArgs{
			"id":                     serverConfig.ID,
			"name":                   serverConfig.Name,
			"root_certificate_id":    serverConfig.RootCertificateID,
			"server_certificate_id":  serverConfig.ServerCertificateID,
			"dh_id":                  serverConfig.DHID,
			"server_config_template": serverConfig.ServerConfigTemplate,
			"client_config_template": serverConfig.ClientConfigTemplate,
			"groups_allowed":         string(gu),
		},
	)

	if err != nil {
		return nil, err
	}

	return &serverConfig, err
}

func (s *ServerProvider) GetServerConfig(id uuid.UUID) (*ServerConfig, error) {
	row := s.postgres.QueryRow(s.ctx, `
			SELECT id, name, root_certificate_id, server_certificate_id, dh_id, server_config_template, client_config_template, groups_allowed
			FROM server_configs WHERE id = $1;`, id)

	serverConfig := &ServerConfig{}
	var groupsAllowed string
	err := row.Scan(
		&serverConfig.ID,
		&serverConfig.Name,
		&serverConfig.RootCertificateID,
		&serverConfig.ServerCertificateID,
		&serverConfig.DHID,
		&serverConfig.ServerConfigTemplate,
		&serverConfig.ClientConfigTemplate,
		&groupsAllowed,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal([]byte(groupsAllowed), &serverConfig.GroupsAllowed)

	serverConfig.provider = s

	return serverConfig, nil
}

func (s *ServerProvider) ListAllowedServers(groups []string) ([]ServerConfigShort, error) {
	servers, err := s.ListServerConfigs()
	if err != nil {
		return nil, err
	}

	var result []ServerConfigShort
	for _, server := range servers {
		for _, group := range groups {
			if slices.Contains(server.GroupsAllowed, group) {
				server.GroupsAllowed = []string{}
				result = append(result, server)
			}
		}
	}

	return result, nil
}

func (s *ServerProvider) ListServerConfigs() ([]ServerConfigShort, error) {
	rows, err := s.postgres.Query(s.ctx, `SELECT id, name, root_certificate_id, server_certificate_id, dh_id, groups_allowed from server_configs;`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var servers []ServerConfigShort
	for rows.Next() {
		server := ServerConfigShort{}
		var groupsAllowed string
		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.RootCertificateID,
			&server.ServerCertificateID,
			&server.DHID,
			&groupsAllowed,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(groupsAllowed), &server.GroupsAllowed)

		servers = append(servers, server)
	}

	return servers, nil
}

func (c *ServerConfig) WriteServerConfigFile(configPath string) error {
	rendered, err := c.GetRendered()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, []byte(rendered), 0600)
}

func (c *ServerConfig) Update(name string, serverCertificateID uuid.UUID, dhID uuid.UUID, serverConfigTemplate string, clientConfigTemplate string, groupsAllowed []string) (bool, error) {
	rootCertificateID := uuid.Nil()

	p := ca.NewProvider(c.provider.ctx, c.provider.postgres)

	rootCert, err := p.GetRootByServerCertificateID(serverCertificateID)
	if err == nil {
		rootCertificateID = rootCert.ID
	}

	trx, err := c.provider.postgres.Begin(c.provider.ctx)
	if err != nil {
		return false, err
	}

	defer func() { _ = trx.Rollback(context.Background()) }()

	row := trx.QueryRow(c.provider.ctx, `
			SELECT id, name, root_certificate_id, server_certificate_id, dh_id, server_config_template, client_config_template, groups_allowed
			FROM server_configs WHERE id=$1 FOR UPDATE`, c.ID,
	)

	updateCache := &ServerConfig{}
	var updateCachegroupsAllowed string
	err = row.Scan(
		&updateCache.ID,
		&updateCache.Name,
		&updateCache.RootCertificateID,
		&updateCache.ServerCertificateID,
		&updateCache.DHID,
		&updateCache.ServerConfigTemplate,
		&updateCache.ClientConfigTemplate,
		&updateCachegroupsAllowed,
	)
	if err != nil {
		return false, err
	}

	changed := false

	if updateCache.Name != name {
		changed = true
		updateCache.Name = name
	}

	if updateCache.ServerCertificateID != serverCertificateID && serverCertificateID != uuid.Nil() {
		changed = true
		updateCache.ServerCertificateID = serverCertificateID
	}

	if updateCache.RootCertificateID != rootCertificateID && rootCertificateID != uuid.Nil() {
		changed = true
		updateCache.RootCertificateID = rootCertificateID
	}

	if updateCache.DHID != dhID && dhID != uuid.Nil() {
		changed = true
		updateCache.DHID = dhID
	}

	if updateCache.ServerConfigTemplate != serverConfigTemplate && serverConfigTemplate != "" {
		changed = true
		updateCache.ServerConfigTemplate = serverConfigTemplate
	}

	if updateCache.ClientConfigTemplate != clientConfigTemplate && clientConfigTemplate != "" {
		changed = true
		updateCache.ClientConfigTemplate = clientConfigTemplate
	}

	gu, err := json.Marshal(groupsAllowed)
	if err != nil {
		return false, err
	}

	if updateCachegroupsAllowed != string(gu) {
		changed = true
	}

	if !changed {
		return false, nil
	}

	_, err = trx.Exec(c.provider.ctx, `
		UPDATE server_configs SET name=@name, root_certificate_id=@root_certificate_id, server_certificate_id=@server_certificate_id, dh_id=@dh_id, server_config_template=@server_config_template, client_config_template=@client_config_template, groups_allowed=@groups_allowed
		WHERE id=@id;`,
		pgx.NamedArgs{
			"id":                     c.ID,
			"name":                   name,
			"root_certificate_id":    updateCache.RootCertificateID,
			"server_certificate_id":  updateCache.ServerCertificateID,
			"dh_id":                  updateCache.DHID,
			"server_config_template": updateCache.ServerConfigTemplate,
			"client_config_template": updateCache.ClientConfigTemplate,
			"groups_allowed":         string(gu),
		},
	)
	if err != nil {
		return false, err
	}

	err = trx.Commit(c.provider.ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *ServerConfig) GetRendered() (string, error) {
	serverConfigBytes := c.ServerConfigTemplate

	tmpl, err := template.New("server").Parse(serverConfigBytes)
	if err != nil {
		return "", err
	}

	certProvider := ca.NewProvider(c.provider.ctx, c.provider.postgres)

	caCert, err := certProvider.GetRootCertificate(c.RootCertificateID)
	if err != nil {
		return "", err
	}
	caCertPEM, err := caCert.GetCertPEM()
	if err != nil {
		return "", err
	}

	crlPEM, err := caCert.GetCRLPEM()
	if err != nil {
		return "", err
	}

	serverCert, err := caCert.GetServerCertificate(c.ServerCertificateID)
	if err != nil {
		return "", err
	}

	serverCertPEM, err := serverCert.GetCertPEM()
	if err != nil {
		return "", err
	}

	serverKeyPEM, err := serverCert.GetKeyPEM()
	if err != nil {
		return "", err
	}

	dhProvider := dh.NewDHProvider(c.provider.ctx, c.provider.postgres)
	dhDer, err := dhProvider.GetDH(c.DHID)
	if err != nil {
		return "", err
	}

	dhPEM, err := dhDer.GetPEM()
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"CA":   caCertPEM,
		"Cert": serverCertPEM,
		"Key":  serverKeyPEM,
		"DH":   dhPEM,
		"CRL":  crlPEM,
	}

	var buffer bytes.Buffer

	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (c *ServerConfig) Delete() error {
	_, err := c.provider.postgres.Exec(c.provider.ctx, `
			DELETE FROM server_configs WHERE id=$1;`, c.ID,
	)

	return err
}

func (c *ServerConfig) ScheduleReload() error {
	_, err := c.provider.postgres.Exec(c.provider.ctx, `
			UPDATE server_states SET reload_scheduled=true WHERE server_id=$1;`, c.ID,
	)

	return err
}
