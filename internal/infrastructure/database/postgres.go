// Package database provides functions to connect to a PostgreSQL database.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/config"
)

// buildDSN creates the Data Source Name string from the config.
func buildDSN(conf *config.Config) string {
	// "postgres://user:password@host:port/dbname?sslmode=disable"
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Name,
		conf.Database.SSLMode,
		conf.Database.TimeZone,
	)
}

// Connect establishes a connection pool to the PostgreSQL database.
func Connect(ctx context.Context, conf *config.Config) (*pgxpool.Pool, error) {
	dsn := buildDSN(conf)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	// Use the parent context for the pool creation itself
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping the database to verify the connection using the provided context
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	return pool, nil
}
