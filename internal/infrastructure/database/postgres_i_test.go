//go:build integration

package database

import (
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/verith/config"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// TestConnect_Integration is an integration test for the Connect function.
func TestConnect_Integration(t *testing.T) {
	// Test Setup
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		t.Fatalf("Could not ping docker: %s", err)
	}

	// Run a PostgreSQL container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_test",
			"POSTGRES_DB=db_test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true // Cleanup after test
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	})

	// Create a config for the test
	port, err := strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		t.Fatalf("Could not get port: %s", err)
	}

	testConfig := &config.Config{
		Database: config.Database{
			User:     "user_test",
			Password: "secret",
			Host:     "localhost",
			Port:     port,
			Name:     "db_test",
			SSLMode:  "disable",
			TimeZone: "UTC",
		},
	}

	// Run the test
	var dbPool *pgxpool.Pool
	// Use an exponential backoff-retry to allow the postgres server to start
	if err := pool.Retry(func() error {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		dbPool, err = Connect(ctx, testConfig)
		return err
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	// Assert
	if dbPool == nil {
		t.Fatalf("Expected dbPool to be initialized, but it is nil")
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		t.Fatalf("Expected dbPool to be connected, but ping failed: %s", err)
	}

	t.Log("Successfully connected to test database and pinged it")
	dbPool.Close() // Close the pool after the test
}
