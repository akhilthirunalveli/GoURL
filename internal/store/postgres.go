package store

import (
	"context"
	"fmt"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(cfg config.DBConfig) (*PostgresStore, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Pool exposes the underlying connection pool (For testing purposes)
func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) SaveLink(link *Link) error {
	query := `INSERT INTO links (original_url, short_code, created_at, expires_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := s.pool.QueryRow(context.Background(), query,
		link.OriginalURL, link.ShortCode, link.CreatedAt, link.ExpiresAt).Scan(&link.ID)

	if err != nil {
		return fmt.Errorf("failed to save link: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetLinkByCode(code string) (*Link, error) {
	query := `SELECT id, original_url, short_code, created_at, expires_at FROM links WHERE short_code = $1`

	var link Link
	err := s.pool.QueryRow(context.Background(), query, code).Scan(
		&link.ID, &link.OriginalURL, &link.ShortCode, &link.CreatedAt, &link.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	return &link, nil
}
