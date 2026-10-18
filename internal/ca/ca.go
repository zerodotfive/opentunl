package ca

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var AlreadyRevoked = errors.New("certificate already revoked")
var InvalidKey = errors.New("invalid key")
var InvalidValidDays = errors.New("valid days can not be negative")

type CertificateProvider struct {
	ctx      context.Context
	postgres *pgxpool.Pool
}

type DefaultCertificate struct {
	Name      string    `json:"name"`
	ID        uuid.UUID `db:"id" json:"id"`
	NotBefore time.Time `db:"not_before" json:"not_before"`
	NotAfter  time.Time `db:"not_after" json:"not_after"`
}

type DefaultCertificateWithRevocation struct {
	DefaultCertificate
	RootID    uuid.UUID `db:"root_id" json:"root_id"`
	IsRevoked bool      `json:"is_revoked"`
}

type RootCertificateShort struct {
	DefaultCertificate
	provider               *CertificateProvider
	ClientDefaultValidDays int `db:"client_default_valid_days" json:"client_default_valid_days"`
}

type RootCertificate struct {
	RootCertificateShort
	Key  []byte `db:"key" json:"key"`
	Cert []byte `db:"cert" json:"cert"`
	CRL  []byte `db:"crl" json:"crl"`
}

type ClientCertificate struct {
	DefaultCertificateWithRevocation
	rootCert *RootCertificate
	Key      []byte `db:"key" json:"key"`
	Cert     []byte `db:"cert" json:"cert"`
}

func NewProvider(ctx context.Context, postgres *pgxpool.Pool) *CertificateProvider {
	return &CertificateProvider{
		ctx:      ctx,
		postgres: postgres,
	}
}

func (p *CertificateProvider) NewRootCertificate(rootName string, validDays int, clientDefaultValidDays int) (*RootCertificate, error) {
	if validDays < 0 {
		return nil, InvalidValidDays
	}

	id := uuid.New()
	timeNow := time.Now()
	notBefore := timeNow
	notAfter := timeNow.Add(time.Duration(validDays) * 24 * time.Hour)

	rootKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	rootKeyBytes, err := x509.MarshalPKCS8PrivateKey(rootKey)
	if err != nil {
		return nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		return nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&rootKey.PublicKey)
	if err != nil {
		return nil, err
	}
	ski := sha1.Sum(pubBytes)

	rootConfig := &x509.Certificate{
		Subject: pkix.Name{
			Country:            []string{"XX"},
			Locality:           []string{"On-Line"},
			Organization:       []string{"VPN"},
			OrganizationalUnit: []string{"CA"},
			CommonName:         rootName,
		},
		SerialNumber:          serial,
		NotBefore:             timeNow,
		NotAfter:              notAfter,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		SubjectKeyId:          ski[:],
	}

	rootCertBytes, err := x509.CreateCertificate(rand.Reader, rootConfig, rootConfig, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}

	crlBytes, err := x509.CreateRevocationList(
		rand.Reader,
		&x509.RevocationList{
			Number:     big.NewInt(1),
			ThisUpdate: timeNow,
			NextUpdate: notAfter,
		},
		rootConfig,
		rootKey,
	)
	if err != nil {
		return nil, err
	}

	_, err = p.postgres.Exec(p.ctx, `
		INSERT INTO root_certificates
		  (id, name, key, cert, crl, client_default_valid_days, not_before, not_after)
		  VALUES
		  (@id, @name, @key, @cert, @crl, @client_default_valid_days, @not_before, @not_after);`,
		pgx.NamedArgs{
			"id":                        id,
			"name":                      rootName,
			"key":                       rootKeyBytes,
			"cert":                      rootCertBytes,
			"crl":                       crlBytes,
			"client_default_valid_days": clientDefaultValidDays,
			"not_before":                notBefore,
			"not_after":                 notAfter,
		},
	)
	if err != nil {
		return nil, err
	}

	return &RootCertificate{
		Name:                   rootName,
		ID:                     id,
		Key:                    rootKeyBytes,
		Cert:                   rootCertBytes,
		CRL:                    crlBytes,
		ClientDefaultValidDays: clientDefaultValidDays,
		NotBefore:              notBefore,
		NotAfter:               notAfter,
	}, nil
}

func (r *RootCertificate) UpdateRootCertificate(clientDefaultValidDays int) (bool, error) {
	if clientDefaultValidDays < 0 {
		return false, InvalidValidDays
	}

	changed := false
	if r.ClientDefaultValidDays != clientDefaultValidDays && clientDefaultValidDays != 0 {
		changed = true
		r.ClientDefaultValidDays = clientDefaultValidDays
	}

	_, err := r.provider.postgres.Exec(r.provider.ctx, `
		UPDATE root_certificates
		SET client_default_valid_days=@client_default_valid_days
		WHERE id=@id;`,
		pgx.NamedArgs{
			"id":                        r.ID,
			"client_default_valid_days": r.ClientDefaultValidDays,
		},
	)
	if err != nil {
		return false, err
	}

	return changed, nil
}

func (p *CertificateProvider) GetRootCertificate(id uuid.UUID) (*RootCertificate, error) {
	row := p.postgres.QueryRow(p.ctx, `SELECT name, id, key, cert, crl, client_default_valid_days, not_before, not_after FROM root_certificates WHERE id=$1;`, id)

	cert := &RootCertificate{}

	err := row.Scan(
		&cert.Name,
		&cert.ID,
		&cert.Key,
		&cert.Cert,
		&cert.CRL,
		&cert.ClientDefaultValidDays,
		&cert.NotBefore,
		&cert.NotAfter,
	)
	if err != nil {
		return nil, err
	}

	cert.provider = p

	return cert, nil
}

func (p *CertificateProvider) ListRootCertificates() ([]RootCertificateShort, error) {
	rows, err := p.postgres.Query(p.ctx, `SELECT name, id, client_default_valid_days, not_before, not_after FROM root_certificates;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []RootCertificateShort
	for rows.Next() {
		cert := RootCertificateShort{}
		err := rows.Scan(
			&cert.Name,
			&cert.ID,
			&cert.ClientDefaultValidDays,
			&cert.NotBefore,
			&cert.NotAfter,
		)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

func (p *CertificateProvider) GetRootByServerCertificateID(id uuid.UUID) (*RootCertificate, error) {
	return p.getRootByCertificateID(id, "server_certificates")
}
func (p *CertificateProvider) GetRootByClientCertificateID(id uuid.UUID) (*RootCertificate, error) {
	return p.getRootByCertificateID(id, "client_certificates")
}

func (p *CertificateProvider) getRootByCertificateID(id uuid.UUID, table string) (*RootCertificate, error) {
	row := p.postgres.QueryRow(p.ctx, fmt.Sprintf(`SELECT root_id FROM %s WHERE id=$1;`, table), id)

	var rootID uuid.UUID
	if err := row.Scan(&rootID); err != nil {
		return nil, err
	}

	return p.GetRootCertificate(rootID)

}

func (r *RootCertificate) GetCRLPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "X509 CRL",
		Bytes: r.CRL,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *RootCertificate) GetCertPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: r.Cert,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *RootCertificate) GetKeyPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: r.Key,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *RootCertificate) Delete() error {
	_, err := r.provider.postgres.Exec(r.provider.ctx, `DELETE FROM root_certificates WHERE id=$1;`, r.ID)

	return err
}

func (r *RootCertificate) NewServerCertificate(OU string, commonName string, validDays int) (*ClientCertificate, error) {
	return r.newCertificate(OU, commonName, "server_certificates", validDays)
}

func (r *RootCertificate) NewClientCertificate(OU string, commonName string, validDays int) (*ClientCertificate, error) {
	return r.newCertificate(OU, commonName, "client_certificates", validDays)
}

func (r *RootCertificate) newCertificate(OU string, commonName string, table string, validDays int) (*ClientCertificate, error) {
	if validDays < 0 {
		return nil, InvalidValidDays
	}

	id := uuid.New()
	timeNow := time.Now()
	notBefore := timeNow
	notAfter := timeNow.Add(time.Duration(validDays) * 24 * time.Hour)

	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		return nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	ski := sha1.Sum(pubBytes)

	var usage []x509.ExtKeyUsage

	if table == "server_certificates" {
		usage = append(usage, x509.ExtKeyUsageServerAuth)
	}

	if table == "client_certificates" {
		usage = append(usage, x509.ExtKeyUsageClientAuth)
	}

	config := &x509.Certificate{
		Subject: pkix.Name{
			Country:            []string{"XX"},
			Locality:           []string{"On-Line"},
			Organization:       []string{"VPN"},
			OrganizationalUnit: []string{OU},
			CommonName:         commonName,
		},
		SerialNumber:          serial,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  false,
		ExtKeyUsage:           usage,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		SubjectKeyId:          ski[:],
	}

	rootKey, err := x509.ParsePKCS8PrivateKey(r.Key)
	if err != nil {
		return nil, err
	}

	rootCert, err := x509.ParseCertificate(r.Cert)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, config, rootCert, &key.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}

	cert := &ClientCertificate{
		ID:        id,
		RootID:    r.ID,
		Key:       keyBytes,
		Cert:      certBytes,
		NotBefore: notBefore,
		NotAfter:  notAfter,
		IsRevoked: false,
	}

	_, err = r.provider.postgres.Exec(r.provider.ctx, fmt.Sprintf(`
		INSERT INTO %s
		  (id, root_id, name, key, cert, not_before, not_after, is_revoked)
		  VALUES
		  (@id, @root_id, @name, @key, @cert, @not_before, @not_after, @is_revoked)`, table),
		pgx.NamedArgs{
			"id":         id,
			"root_id":    r.ID,
			"name":       commonName,
			"key":        keyBytes,
			"cert":       certBytes,
			"not_before": notBefore,
			"not_after":  notAfter,
			"is_revoked": false,
		},
	)

	cert.rootCert = r
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (r *RootCertificate) GetServerCertificate(id uuid.UUID) (*ClientCertificate, error) {
	return r.getCertificate(id, "server_certificates")
}

func (r *RootCertificate) GetClientCertificate(id uuid.UUID) (*ClientCertificate, error) {
	return r.getCertificate(id, "client_certificates")
}

func (r *RootCertificate) getCertificate(id uuid.UUID, table string) (*ClientCertificate, error) {
	row := r.provider.postgres.QueryRow(r.provider.ctx, fmt.Sprintf(`
				SELECT id, root_id, name, key, cert, not_before, not_after, is_revoked
				FROM %s WHERE id=$1 AND root_id=$2;`, table),
		id, r.ID,
	)

	cert := &ClientCertificate{}

	err := row.Scan(
		&cert.ID,
		&cert.RootID,
		&cert.Name,
		&cert.Key,
		&cert.Cert,
		&cert.NotBefore,
		&cert.NotAfter,
		&cert.IsRevoked,
	)

	cert.rootCert = r

	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (r *RootCertificate) GetCountValidClientCertificatesByName(name string) (int, error) {
	var count int

	err := r.provider.postgres.QueryRow(r.provider.ctx, `
		SELECT COUNT(*)
		FROM client_certificates
		WHERE root_id = @root_id
		  AND name = @name
		  AND is_revoked = false
		  AND not_before < NOW();`,
		pgx.NamedArgs{
			"root_id": r.ID,
			"name":    name,
		},
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *RootCertificate) GetValidClientCertificateByName(name string) (*ClientCertificate, error) {
	row := r.provider.postgres.QueryRow(r.provider.ctx, `
		SELECT name, id, root_id, key, cert, not_before, not_after, is_revoked
		FROM client_certificates
		WHERE root_id = @root_id
		  AND name = @name
		  AND is_revoked = false
		  AND not_before < NOW()
		ORDER BY not_before DESC
		LIMIT 1;`,
		pgx.NamedArgs{
			"root_id": r.ID,
			"name":    name,
		},
	)

	cert := &ClientCertificate{}

	err := row.Scan(
		&cert.Name,
		&cert.ID,
		&cert.RootID,
		&cert.Key,
		&cert.Cert,
		&cert.NotBefore,
		&cert.NotAfter,
		&cert.IsRevoked,
	)

	cert.rootCert = r

	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (r *RootCertificate) ListServerCertificates() ([]DefaultCertificateWithRevocation, error) {
	return r.listCertificates("server_certificates")
}

func (r *RootCertificate) ListClientCertificates() ([]DefaultCertificateWithRevocation, error) {
	return r.listCertificates("client_certificates")
}

func (r *RootCertificate) listCertificates(table string) ([]DefaultCertificateWithRevocation, error) {
	rows, err := r.provider.postgres.Query(r.provider.ctx, fmt.Sprintf(`SELECT name, id, root_id, not_before, not_after, is_revoked FROM %s WHERE root_id=$1;`, table), r.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var certs []DefaultCertificateWithRevocation
	for rows.Next() {
		cert := DefaultCertificateWithRevocation{}
		err := rows.Scan(
			&cert.Name,
			&cert.ID,
			&cert.RootID,
			&cert.NotBefore,
			&cert.NotAfter,
			&cert.IsRevoked,
		)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

func (c *ClientCertificate) RevokeServerCertificate(id uuid.UUID) error {
	return c.revokeCertificate(id, "server_certificates")
}

func (c *ClientCertificate) RevokeClientCertificate(id uuid.UUID) error {
	return c.revokeCertificate(id, "client_certificates")
}

func (c *ClientCertificate) revokeCertificate(id uuid.UUID, table string) error {
	if c.IsRevoked {
		return AlreadyRevoked
	}

	timeNow := time.Now()

	cert, err := x509.ParseCertificate(c.Cert)
	if err != nil {
		return err
	}

	rootKey, err := x509.ParsePKCS8PrivateKey(c.rootCert.Key)
	if err != nil {
		return err
	}

	rootCert, err := x509.ParseCertificate(c.rootCert.Cert)
	if err != nil {
		return err
	}

	crl, err := x509.ParseRevocationList(c.rootCert.CRL)
	if err != nil {
		return err
	}

	entries := append(
		append([]x509.RevocationListEntry{}, crl.RevokedCertificateEntries...),
		x509.RevocationListEntry{
			SerialNumber:   cert.SerialNumber,
			RevocationTime: timeNow,
			ReasonCode:     0,
		},
	)

	ck, ok := rootKey.(*ecdsa.PrivateKey)
	if !ok {
		return InvalidKey
	}

	crlBytes, err := x509.CreateRevocationList(
		rand.Reader,
		&x509.RevocationList{
			Number:                    new(big.Int).Add(crl.Number, big.NewInt(1)),
			ThisUpdate:                timeNow,
			NextUpdate:                c.rootCert.NotAfter,
			RevokedCertificateEntries: entries,
		},
		rootCert,
		ck,
	)

	if err != nil {
		return err
	}

	trx, err := c.rootCert.provider.postgres.Begin(c.rootCert.provider.ctx)
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback(c.rootCert.provider.ctx) }()

	_, err = trx.Exec(c.rootCert.provider.ctx, `UPDATE root_certificates SET crl=$1 WHERE id=$2;`, crlBytes, c.RootID)
	if err != nil {
		return err
	}

	_, err = trx.Exec(c.rootCert.provider.ctx, fmt.Sprintf(`UPDATE %s SET is_revoked=true WHERE id=$1;`, table), id)
	if err != nil {
		return err
	}

	return trx.Commit(c.rootCert.provider.ctx)
}

func (c *ClientCertificate) GetCertPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.Cert,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *ClientCertificate) GetKeyPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: c.Key,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
