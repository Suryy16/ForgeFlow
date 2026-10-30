
package task

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	return db
}

func TestGetTask_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	_, err := repo.GetByID(
		context.Background(),
		uuid.New(),
	)

	if err != ErrTaskNotFound {
		t.Fatalf(
			"expected ErrTaskNotFound, got %v",
			err,
		)
	}
}
