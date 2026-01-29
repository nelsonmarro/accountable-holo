// Package database provides functions to connect to a PostgreSQL database.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/config"
	"github.com/nelsonmarro/verith/migrations"
)

// buildDSN creates the Data Source Name string from the config.
// It explicitly sets search_path=public to avoid migration table conflicts.
func buildDSN(conf *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s&search_path=public",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Name,
		conf.Database.SSLMode,
		conf.Database.TimeZone,
	)
}

// RunMigrations applies database migrations automatically.
func RunMigrations(pool *pgxpool.Pool) error {
	_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Prepare migration source from embedded files
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	// 2. Initialize migration engine
	// Note: Connect string includes search_path=public for reliability
	dsn := pool.Config().ConnString()
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// 3. Execute migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// Connect establishes a connection pool to the PostgreSQL database.
func Connect(ctx context.Context, conf *config.Config) (*pgxpool.Pool, error) {
	dsn := buildDSN(conf)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the database config: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return pool, nil
}
