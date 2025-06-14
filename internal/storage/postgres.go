package storage

import (
	"context"
	"errors"
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

func (r PgRepository) Add(short, original string, userID int64) error {
	row := r.pool.QueryRow(context.Background(), `
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

func (r PgRepository) AddBatch(ctx context.Context, userID int64, b map[string]string) error {
	batch := &pgx.Batch{}
	for short, original := range b {
		batch.Queue("INSERT INTO url (short, original, user_id) VALUES ($1, $2, $3)", short, original, userID)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := range len(b) {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("error executing batch command %d: %w", i, err)
		}
	}
	return nil
}

func (r PgRepository) GetUserURLs(ctx context.Context, userID int64) ([]StoredURL, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT short, original
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
		if err := rows.Scan(&url.ShortID, &url.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (r PgRepository) MarkDeletedUserURLs(ctx context.Context, userID int64, shortIDs ...string) {
	batch := &pgx.Batch{}
	for _, shortID := range shortIDs {
		batch.Queue("UPDATE url SET is_deleted=true WHERE short=$1 AND user_id=$2", shortID, userID)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()
	for i := range len(shortIDs) {
		_, err := results.Exec()
		if err != nil {
			fmt.Printf("error executing batch command %d: %v", i, err)
		}
	}
}

//func (r PgRepository) SaveMessages(ctx context.Context, messages ...StoredURL) error {
//	// соберём данные для создания запроса с групповой вставкой
//	var values []string
//	var args []any
//	for i, msg := range messages {
//		// в нашем запросе по 4 параметра на каждое сообщение
//		base := i * 4
//		// PostgreSQL требует шаблоны в формате ($1, $2, $3, $4) для каждой вставки
//		params := fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
//		values = append(values, params)
//		args = append(args, msg.Sender, msg.Recepient, msg.Payload, msg.Time)
//	}
//
//	// составляем строку запроса
//	query := `
//  INSERT INTO url
//  (sender, recepient, payload, sent_at)
//  VALUES ` + strings.Join(values, ",") + `;`
//
//	// добавляем новые сообщения в БД
//	_, err := s.conn.ExecContext(ctx, query, args...)
//
//	return err
//}
