package node

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeProvider struct {
	ctx      context.Context
	postgres *pgxpool.Pool
}

type Node struct {
	Name     string    `db:"name" json:"name"`
	LastSeen time.Time `db:"last_seen" json:"last_seen"`
	Online   bool      `json:"online"`
}

func NewNodeProvider(ctx context.Context, postgres *pgxpool.Pool) *NodeProvider {
	return &NodeProvider{
		ctx:      ctx,
		postgres: postgres,
	}
}

func (p *NodeProvider) ListNodes() ([]Node, error) {
	rows, err := p.postgres.Query(p.ctx, `SELECT name, last_seen FROM nodes;`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.Name, &n.LastSeen); err != nil {
			return nil, err
		}
		n.Online = n.LastSeen.After(time.Now().Add(-time.Minute))
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (p *NodeProvider) DeleteNode(name string) error {
	_, err := p.postgres.Exec(p.ctx, `DELETE FROM nodes WHERE name = $1;`, name)
	return err
}

func (p *NodeProvider) heartbeat(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := p.postgres.Exec(ctx, `
				INSERT INTO nodes (name, last_seen)
				VALUES ($1, now())
				ON CONFLICT (name) DO UPDATE SET last_seen = now();`,
		name,
	)

	return err
}

func (p *NodeProvider) Start() {
	go func() {
		name, err := os.Hostname()
		if err != nil {
			log.Fatal("heartbeat get hostname error: %s", err.Error())
		}
		for _ = range time.NewTicker(10 * time.Second).C {
			if err := p.heartbeat(name); err != nil {
				log.Printf("heartbeat write error: %s", err.Error())
			}
		}
	}()
}
