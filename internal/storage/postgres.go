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
