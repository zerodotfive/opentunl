package ca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
	"uuid"
)

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
