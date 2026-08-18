package backend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Record is one saved evaluation: the raw input, a headline number, and a label.
// This shape is generic across the slate's apps so the persistence layer is shared.
type Record struct {
	ID        int64           `json:"id"`
	CreatedAt time.Time       `json:"createdAt"`
	Input     json.RawMessage `json:"input"`
	Headline  float64         `json:"headline"`
	Label     string          `json:"label"`
}

// Store persists and lists records.
type Store interface {
	Save(rec Record) (Record, error)
	List(limit int) ([]Record, error)
}

// PostgresStore is a Postgres-backed Store.
type PostgresStore struct{ pool *pgxpool.Pool }

const schemaDDL = `
CREATE TABLE IF NOT EXISTS records (
	id         BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	input      JSONB NOT NULL,
	headline   DOUBLE PRECISION NOT NULL,
	label      TEXT NOT NULL DEFAULT ''
);`

// NewPostgresStore connects, pings, and ensures the schema.
func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, schemaDDL); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

// Close releases the pool.
func (s *PostgresStore) Close() { s.pool.Close() }

// Save inserts one record and returns it with id/created_at populated.
func (s *PostgresStore) Save(rec Record) (Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx,
		`INSERT INTO records (input, headline, label) VALUES ($1, $2, $3) RETURNING id, created_at`,
		[]byte(rec.Input), rec.Headline, rec.Label)
	if err := row.Scan(&rec.ID, &rec.CreatedAt); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// List returns the most recent records, newest first.
func (s *PostgresStore) List(limit int) ([]Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx,
		`SELECT id, created_at, input, headline, label FROM records ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Record
	for rows.Next() {
		var rec Record
		var raw []byte
		if err := rows.Scan(&rec.ID, &rec.CreatedAt, &raw, &rec.Headline, &rec.Label); err != nil {
			return nil, err
		}
		rec.Input = json.RawMessage(raw)
		out = append(out, rec)
	}
	return out, rows.Err()
}
