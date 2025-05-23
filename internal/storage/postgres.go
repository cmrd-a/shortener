package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Check(dsn string) error {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	_, err = conn.Query(context.Background(), "SELECT 1")
	if err != nil {
		return err
	}
	return nil
}

type PgRepository struct {
	dsn string
}

func NewPgRepository(dsn string) (*PgRepository, error) {
	r := &PgRepository{dsn: dsn}
	err := r.MakeQuery(`
		CREATE TABLE IF NOT EXISTS url
		(
			id       BIGSERIAL,
			short    text NOT NULL UNIQUE,
			original text NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}
	err = r.MakeQuery(`
	CREATE UNIQUE INDEX IF NOT EXISTS urls_short_uindex
    ON url (short)
	`)
	if err != nil {
		return nil, err
	}

	return r, nil
}
func (r PgRepository) MakeQuery(sql string, args ...any) error {
	conn, err := pgx.Connect(context.Background(), r.dsn)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	_, err = conn.Query(context.Background(), sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r PgRepository) Get(short string) (string, error) {
	conn, err := pgx.Connect(context.Background(), r.dsn)
	if err != nil {
		return "", err
	}
	defer conn.Close(context.Background())

	var original string
	err = conn.QueryRow(context.Background(), "SELECT original FROM url  WHERE short=$1", short).Scan(&original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (r PgRepository) Add(short, original string) error {
	err := r.MakeQuery("INSERT INTO url (short, original) VALUES ($1, $2)", short, original)
	if err != nil {
		return err
	}
	return nil
}
