package ovpn

import (
	"bytes"
	"errors"
	"slices"
	"text/template"

	"github.com/jackc/pgx/v5"
	"github.com/zerodotfive/opentunl/internal/ca"
)

var ClientIsNotAllowed = errors.New("client is not allowed by server")
var MaxCertsReached = errors.New("reached maximum valid certificates allowed for client")

func (c *ServerConfig) GetClientConfigRenderedClient(name string, groups []string, forceNewCert bool, withLimit bool) (string, error) {
	allowed := false
	for _, group := range groups {
		if slices.Contains(c.GroupsAllowed, group) {
			allowed = true
		}
	}
	if !allowed {
		return "", ClientIsNotAllowed
	}

	return c.GetClientConfigRendered(name, 0, forceNewCert, withLimit)
}

func (c *ServerConfig) GetClientConfigRendered(name string, validDays int, forceNewCert bool, withLimit bool) (string, error) {
	vd := validDays

	tmpl, err := template.New("config").Parse(c.ClientConfigTemplate)
	if err != nil {
		return "", err
	}

	provider := ca.NewProvider(c.provider.ctx, c.provider.postgres)

	caCert, err := provider.GetRootCertificate(c.RootCertificateID)
	if err != nil {
		return "", err
	}

	if vd == 0 {
		vd = caCert.ClientDefaultValidDays
	}

	caCertPEM, err := caCert.GetCertPEM()
	if err != nil {
		return "", err
	}

	count, err := caCert.GetCountValidClientCertificatesByName(name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	clientCert, err := caCert.GetValidClientCertificateByName(name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	if forceNewCert {
		if count >= 2 && withLimit {
			return "", MaxCertsReached
		}

		clientCert = nil
	}

	if clientCert == nil {
		clientCert, err = caCert.NewClientCertificate(name, name, vd)
		if err != nil {
			return "", err
		}
	}

	clientCertPEM, err := clientCert.GetCertPEM()
	if err != nil {
		return "", err
	}

	clientKeyPEM, err := clientCert.GetKeyPEM()
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"CA":   caCertPEM,
		"Cert": clientCertPEM,
		"Key":  clientKeyPEM,
	}

	var buffer bytes.Buffer

	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
