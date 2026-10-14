package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Suryy16/ForgeFlow/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/joho/godotenv"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const usage = "usage: go run ./cmd/migrate <up|down|version|force <version>>"

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) (err error) {
	if len(args) < 1 {
		return errors.New(usage)
	}

	// Abaikan error: di production env biasanya berasal dari sistem, bukan file.
	_ = godotenv.Load()

	databaseURL, _ := config.DatabaseURL()
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("initialize migration: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			log.Printf("close migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("close migration database: %v", databaseErr)
		}
	}()

	switch args[0] {
	case "up":
		err = m.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Database is already up to date.")
			return nil
		}

	case "down":
		err = m.Steps(-1)
		if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, os.ErrNotExist) {
			fmt.Println("No migration to roll back.")
			return nil
		}

	case "version":
		version, dirty, versionErr := m.Version()
		if errors.Is(versionErr, migrate.ErrNilVersion) {
			fmt.Println("Version: none (no migration applied yet)")
			return nil
		}
		if versionErr != nil {
			return fmt.Errorf("get migration version: %w", versionErr)
		}
		fmt.Printf("Version: %d, Dirty: %t\n", version, dirty)
		return nil

	case "force":
		if len(args) < 2 {
			return errors.New("usage: go run ./cmd/migrate force <version>")
		}
		v, convErr := strconv.Atoi(args[1])
		if convErr != nil {
			return fmt.Errorf("invalid version %q: %w", args[1], convErr)
		}
		if err = m.Force(v); err != nil {
			return fmt.Errorf("force version: %w", err)
		}
		fmt.Printf("Forced version to %d.\n", v)
		return nil

	default:
		return fmt.Errorf("unknown command %q; %s", args[0], usage)
	}

	var dirtyErr migrate.ErrDirty
	if errors.As(err, &dirtyErr) {
		return fmt.Errorf("database is dirty at version %d: fix the SQL, then run `force <version>`", dirtyErr.Version)
	}
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migration completed successfully.")
	return nil
}
