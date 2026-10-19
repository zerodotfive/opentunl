package ca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/jackc/pgx/v5"
	"math/big"
	"time"
	"uuid"
)

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
