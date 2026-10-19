package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/jackc/pgx/v5"
	"math/big"
	"time"
	"uuid"
)

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
