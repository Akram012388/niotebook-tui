package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBURL() string {
	if url := os.Getenv("NIOTEBOOK_TEST_DB_URL"); url != "" {
		return url
	}
	return "postgres://localhost/niotebook_test?sslmode=disable"
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := testDBURL()

	// Run migrations
	m, err := migrate.New("file://../../../migrations", dbURL)
	if err != nil {
		t.Fatalf("migrate new: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, posts, refresh_tokens CASCADE")
		pool.Close()
	})

	return pool
}
