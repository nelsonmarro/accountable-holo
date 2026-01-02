//go:build integration

package persistence

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import the postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Import the file source driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

var (
	testRepo       *AccountRepositoryImpl
	testCatRepo    *CategoryRepositoryImpl
	testReportRepo *ReportRepositoryImpl
	testUserRepo   *UserRepositoryImpl
	dbPool         *pgxpool.Pool // Make the pool accessible to helpers
	testUser       *domain.User  // Global test user
)

// TestMain is the entry point for all tests in this package.
func TestMain(m *testing.M) {
	// --- Setup: Start the PostgreSQL container ---
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}
	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_test",
			"POSTGRES_DB=db_test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// --- Connect to the database ---
	var databaseURL string
	if err := pool.Retry(func() error {
		var err error
		databaseURL = fmt.Sprintf("postgres://user_test:secret@localhost:%s/db_test?sslmode=disable", resource.GetPort("5432/tcp"))
		tempPool, err := pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			return err
		}
		defer tempPool.Close()
		return tempPool.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to database after retries: %s", err)
	}

	// --- Run Migrations using golang-migrate ---
	migrationPath := "file://../../../migrations"

	mig, err := migrate.New(migrationPath, databaseURL)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %s", err)
	}
	// Apply all "up" migrations. The `err != migrate.ErrNoChange` part handles
	// the case where migrations have already been applied.
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run up migrations: %s", err)
	}
	log.Println("Migrations applied successfully!")

	// Create the final connection pool for tests.
	dbPool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Could not create final connection pool: %v", err)
	}

	// Create the repository instance that all tests will use.
	testRepo = NewAccountRepository(dbPool)
	testCatRepo = NewCategoryRepository(dbPool)
	testReportRepo = NewReportRepository(dbPool)
	testUserRepo = NewUserRepository(dbPool)

	// --- Run the tests ---
	code := m.Run()

	// --- Teardown: Clean up the container ---
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// truncateTables cleans the database tables between test runs for isolation.
func truncateTables(t *testing.T) {
	_, err := dbPool.Exec(context.Background(), "TRUNCATE TABLE accounts, categories, transactions, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}
}

// Helper function to create a test user
func createTestUser(t *testing.T, repo *UserRepositoryImpl, username string, role domain.UserRole) *domain.User {
	user := &domain.User{
		Username:     username,
		PasswordHash: "hashed_password",
		FirstName:    "Test",
		LastName:     "User",
		Role:         role,
	}
	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err, "Failed to create test user")
	return user
}
