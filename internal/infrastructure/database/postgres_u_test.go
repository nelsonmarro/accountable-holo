package database

import (
	"testing"

	"github.com/nelsonmarro/accountable-holo/config"
)

// TestBuildDSN is a unit test for the DSN fomatting logic.
func TestBuildDSN(t *testing.T) {
	t.Run("should correctly format DSN string", func(t *testing.T) {
		// ArrangeL Create a sample config
		conf := &config.Config{
			Database: config.Database{
				User:     "testuser",
				Password: "testpass",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				SSLMode:  "disable",
				TimeZone: "UTC",
			},
		}

		expectedDSN := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable&TimeZone=UTC"

		// Act: Call the buildDSN function
		dsn := buildDSN(conf)

		// Assert: Check if the formatted DSN matches the expected value
		if dsn != expectedDSN {
			t.Errorf("expected DSN '%s', got '%s'", expectedDSN, dsn)
		}
	})
}
