package ca

import (
	"context"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
