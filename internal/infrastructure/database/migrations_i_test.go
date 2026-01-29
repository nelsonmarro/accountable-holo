package database

import (
	"context"
	"testing"
	"time"

	"github.com/nelsonmarro/verith/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Load configuration (pointing to test database)
	conf, err := config.LoadConfig("../../../config")
	require.NoError(t, err)
	
	// Override database name for safety if needed, 
	// but usually developers use a dedicated 'verith_test' DB in their local env.
	// For this test, we assume the environment is set up or the default test DB is used.

	// 2. Connect to the database
	pool, err := Connect(ctx, conf)
	require.NoError(t, err)
	defer pool.Close()

	// 3. Run migrations
	err = RunMigrations(pool)
	assert.NoError(t, err, "RunMigrations should apply all scripts without error")

	// 4. Verification: Check if a table from the migrations exists
	var exists bool
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'issuers')"
	err = pool.QueryRow(ctx, query).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "Table 'issuers' should exist after migrations")

	// 5. Verification: Check if the NEW column exists (from the latest migration)
	var columnExists bool
	colQuery := "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'issuers' AND column_name = 'default_tax_rate')"
	err = pool.QueryRow(ctx, colQuery).Scan(&columnExists)
	require.NoError(t, err)
	assert.True(t, columnExists, "Column 'default_tax_rate' should exist in 'issuers' table")
	
	// 6. Check users table (fundamental)
	err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
	assert.True(t, exists, "Table 'users' should exist")
}

func TestRunMigrations_Idempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conf, _ := config.LoadConfig("../../../config")
	pool, err := Connect(ctx, conf)
	require.NoError(t, err)
	defer pool.Close()

	// 1. First run (already covered by main test, but we do it here for isolation)
	err = RunMigrations(pool)
	assert.NoError(t, err)

	// 2. Second run immediately after
	err = RunMigrations(pool)
	assert.NoError(t, err, "Running migrations a second time should not return error (migrate.ErrNoChange should be handled)")
}
