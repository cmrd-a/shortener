package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository implements the Repository interface using PostgreSQL as the storage backend.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PgRepository instance with a PostgreSQL connection pool.
// It initializes the database schema by calling Bootstrap().
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

// Bootstrap creates the necessary database tables and indexes for the URL shortener.
func (r PgRepository) Bootstrap() error {
	_, err := r.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS url
		(
			id       BIGSERIAL PRIMARY KEY,
			user_id  BIGINT NOT NULL DEFAULT 0,
			short    text NOT NULL,
			original text NOT NULL,
			is_deleted bool NOT NULL DEFAULT false,
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
	_, err = r.pool.Exec(context.Background(), `
		CREATE UNIQUE INDEX IF NOT EXISTS urls_original_uindex
		ON url (original)
	`)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(context.Background(), `
		CREATE INDEX IF NOT EXISTS user_id_index
		ON url (user_id)
	`)
	if err != nil {
		return err
	}
	return nil
}

// Ping checks the health of the PostgreSQL database connection.
func (r PgRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Get retrieves the original URL for a given short URL identifier from PostgreSQL.
func (r PgRepository) Get(ctx context.Context, short string) (string, error) {
	var original string
	var isDeleted bool
	err := r.pool.QueryRow(ctx, "SELECT original, is_deleted FROM url  WHERE short=$1", short).Scan(&original, &isDeleted)
	if err != nil {
		return "", err
	}
	if isDeleted {
		return "", ErrURLIsDeleted
	}
	return original, nil
}

// Add stores a new URL mapping in PostgreSQL, checking for duplicates.
func (r PgRepository) Add(ctx context.Context, short, original string, userID int64) error {
	row := r.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO url
				(user_id, short, original)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING),
			 dup AS (SELECT short
					 FROM url
					 WHERE original = $3)
		SELECT short
		FROM dup
	`, userID, short, original)
	var existingShort string
	err := row.Scan(&existingShort)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return NewOriginalExistError(existingShort)
}

// AddBatch stores multiple URL mappings in PostgreSQL using a batch operation for efficiency.
func (r PgRepository) AddBatch(ctx context.Context, userID int64, batch ...StoredURL) error {
	b := &pgx.Batch{}
	for _, url := range batch {
		b.Queue("INSERT INTO url (short, original, user_id) VALUES ($1, $2, $3)", url.ShortID, url.OriginalURL, userID)
	}
	results := r.pool.SendBatch(ctx, b)
	defer results.Close()

	for i := range len(batch) {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("error executing batch command %d: %w", i, err)
		}
	}
	return nil
}

// GetUserURLs retrieves all non-deleted URLs created by a specific user from PostgreSQL.
func (r PgRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT short, original, is_deleted
		FROM url
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls = make([]StoredURL, 0)
	for rows.Next() {
		url := StoredURL{}
		if err := rows.Scan(&url.ShortID, &url.OriginalURL, &url.IsDeleted); err != nil {
			return nil, err
		}
		if !url.IsDeleted {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

// MarkDeletedUserURLs marks the specified URLs as deleted in PostgreSQL using a batch operation.
func (r PgRepository) MarkDeletedUserURLs(ctx context.Context, urls ...URLForDelete) {
	batch := &pgx.Batch{}
	for _, url := range urls {
		batch.Queue("UPDATE url SET is_deleted=true WHERE short=$1 AND user_id=$2", url.ShortID, url.UserID)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()
	for i := range len(urls) {
		_, err := results.Exec()
		if err != nil {
			fmt.Printf("error executing batch command %d: %v", i, err)
		}
	}
}
