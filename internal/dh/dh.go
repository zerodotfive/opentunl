package dh

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DHProvider struct {
	ctx      context.Context
	postgres *pgxpool.Pool
}

type DHShort struct {
	provider *DHProvider
	ID       uuid.UUID `db:"id" json:"id"`
	Name     string    `db:"name" json:"name"`
}
type DH struct {
	DHShort
	DerBytes []byte `db:"der_bytes" json:"der_bytes"`
}

type dhParameters struct {
	Prime *big.Int
	Base  *big.Int
}

func NewDHProvider(ctx context.Context, postgres *pgxpool.Pool) *DHProvider {
	return &DHProvider{
		ctx:      ctx,
		postgres: postgres,
	}
}

func (p *DHProvider) NewDH(name string) error {
	id := uuid.New()
	r, err := rand.Prime(rand.Reader, 4096)
	if err != nil {
		return err
	}

	g := big.NewInt(2)

	derBytes, err := asn1.Marshal(dhParameters{
		Prime: r,
		Base:  g,
	})
	if err != nil {
		return err
	}

	_, err = p.postgres.Exec(p.ctx, `INSERT INTO dh (id, name, der_bytes) VALUES ($1, $2, $3);`, id, name, derBytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *DHProvider) GetDH(id uuid.UUID) (*DH, error) {
	row := p.postgres.QueryRow(p.ctx, `SELECT id, name, der_bytes FROM dh WHERE id = $1;`, id)

	result := DH{}

	err := row.Scan(&result.ID, &result.Name, &result.DerBytes)
	if err != nil {
		return nil, err
	}

	result.provider = p

	return &result, nil
}

func (p *DHProvider) ListDH() ([]DHShort, error) {
	rows, err := p.postgres.Query(p.ctx, `SELECT name, id from dh;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ds []DHShort
	for rows.Next() {
		d := DHShort{}
		err := rows.Scan(
			&d.Name,
			&d.ID,
		)
		if err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}

	return ds, nil
}

func (d *DH) Delete() error {
	_, err := d.provider.postgres.Exec(d.provider.ctx, `DELETE FROM dh WHERE id = $1;`, d.ID)

	return err
}

func (d *DH) GetPEM() (string, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{
		Type:  "DH PARAMETERS",
		Bytes: d.DerBytes,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
