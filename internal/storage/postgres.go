package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(dsn string) (*PgRepository, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	r := &PgRepository{pool: pool}
	err = r.Bootstrap()
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (r PgRepository) Bootstrap() error {
	_, err := r.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS url
		(
			id       BIGSERIAL PRIMARY KEY,
			short    text NOT NULL UNIQUE,
			original text NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(context.Background(), `
		CREATE UNIQUE INDEX IF NOT EXISTS urls_short_uindex
		ON url (short)
	`)
	if err != nil {
		return err
	}
	return nil
}

func (r PgRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r PgRepository) Get(short string) (string, error) {
	var original string
	err := r.pool.QueryRow(context.Background(), "SELECT original FROM url  WHERE short=$1", short).Scan(&original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (r PgRepository) Add(short, original string) error {
	_, err := r.pool.Exec(context.Background(), "INSERT INTO url (short, original) VALUES ($1, $2)", short, original)
	if err != nil {
		return err
	}
	return nil
}

func (r PgRepository) AddBatch(ctx context.Context, b map[string]string) error {
	batch := &pgx.Batch{}
	for short, original := range b {
		batch.Queue("INSERT INTO url (short, original) VALUES ($1, $2)", short, original)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(b); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("error executing batch command %d: %w", i, err)
		}
	}
	return nil
}
