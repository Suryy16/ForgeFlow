package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(dsn string) (*pgxpool.Pool, error) {
	// Parse konfigurasi pool jika ingin melakukan kustomisasi lebih lanjut
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Mengatur konfigurasi connection pool (menggantikan SetMaxOpenConns, dll.)
	config.MaxConns = 10
	config.MaxConnLifetime = 30 * time.Minute
	// Anda juga bisa mengatur MaxConnIdleTime, HealthCheckPeriod, dll. jika diperlukan
	// config.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Membuat pool koneksi dan sekaligus melakukan ping ke database
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect pool: %w", err)
	}

	// Pastikan koneksi benar-benar siap
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Jangan lupa tutup pool jika ping gagal agar tidak leak
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
