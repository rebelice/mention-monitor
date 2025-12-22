package notifier

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rebelice/mention-monitor/internal/models"
)

// Postgres stores mentions in PostgreSQL (Supabase)
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new Postgres notifier
func NewPostgres(ctx context.Context, connString string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Create table if not exists
	if err := createTable(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func createTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS mentions (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			type TEXT NOT NULL,
			keyword TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			url TEXT NOT NULL,
			author TEXT,
			discovered_at TIMESTAMPTZ NOT NULL,
			published_at TIMESTAMPTZ,
			status TEXT DEFAULT 'unread',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_mentions_discovered_at ON mentions(discovered_at DESC);
		CREATE INDEX IF NOT EXISTS idx_mentions_url ON mentions(url);
	`
	_, err := pool.Exec(ctx, query)
	return err
}

// Send stores mentions in PostgreSQL
func (p *Postgres) Send(ctx context.Context, mentions []models.Mention) error {
	if len(mentions) == 0 {
		return nil
	}

	for _, m := range mentions {
		err := p.insertMention(ctx, m)
		if err != nil {
			fmt.Printf("Failed to insert mention %s: %v\n", m.ID, err)
		}
	}

	return nil
}

func (p *Postgres) insertMention(ctx context.Context, m models.Mention) error {
	query := `
		INSERT INTO mentions (id, source, type, keyword, title, content, url, author, discovered_at, published_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'unread', NOW())
		ON CONFLICT (id) DO NOTHING
	`

	_, err := p.pool.Exec(ctx, query,
		m.ID,
		m.Source,
		m.Type,
		m.Keyword,
		m.Title,
		m.Content,
		m.URL,
		m.Author,
		m.DiscoveredAt,
		m.PublishedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// CheckDuplicate checks if a mention with the given ID already exists
func (p *Postgres) CheckDuplicate(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM mentions WHERE id = $1)`
	err := p.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Close closes the PostgreSQL connection pool
func (p *Postgres) Close() {
	p.pool.Close()
}
