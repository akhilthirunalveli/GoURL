package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/models"
	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func New(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) CreateURL(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := d.db.QueryRowContext(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
		time.Now(),
		time.Now(),
	).Scan(&url.ID)

	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (d *Database) GetURLByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, updated_at, click_count
		FROM urls
		WHERE short_code = $1
	`

	url := &models.URL{}
	err := d.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return url, nil
}

func (d *Database) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `
		UPDATE urls
		SET click_count = click_count + 1, updated_at = $1
		WHERE short_code = $2
	`

	_, err := d.db.ExecContext(ctx, query, time.Now(), shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}

func (d *Database) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	var exists bool
	err := d.db.QueryRowContext(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	return exists, nil
}
