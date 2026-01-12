package licensing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseManager_CheckStatus(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "license_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	manager := NewLicenseManager(tmpDir)

	t.Run("New Install Starts Trial", func(t *testing.T) {
		status, err := manager.CheckStatus()
		require.NoError(t, err)
		require.NotNil(t, status)
		
		assert.Equal(t, StatusTrial, status.Status)
		// Should be today
		assert.WithinDuration(t, time.Now(), status.InstallDate, 1*time.Minute)
	})

	t.Run("Active License Remains Active", func(t *testing.T) {
		// Manually create an active license file
		activeData := &LicenseData{
			InstallDate: time.Now().AddDate(0, -1, 0), // Installed a month ago
			Status:      StatusActive,
			LicenseKey:  "TEST-KEY",
		}
		require.NoError(t, manager.store.Save(activeData))

		status, err := manager.CheckStatus()
		require.NoError(t, err)
		assert.Equal(t, StatusActive, status.Status)
	})

	t.Run("Trial Expires After 15 Days", func(t *testing.T) {
		// Manually create an old trial file (16 days ago)
		expiredData := &LicenseData{
			InstallDate: time.Now().AddDate(0, 0, -16),
			Status:      StatusTrial,
		}
		require.NoError(t, manager.store.Save(expiredData))

		status, err := manager.CheckStatus()
		require.NoError(t, err)
		assert.Equal(t, StatusExpired, status.Status)
	})
}

func TestActivateLicense_Integration(t *testing.T) {
	// Setup Mock Server simulating Lemon Squeezy API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/licenses/activate", r.URL.Path)

		// Decode payload
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)

		w.Header().Set("Content-Type", "application/json")

		// Simulate Success for a specific key
		if payload["license_key"] == "VALID-KEY" {
			_ = json.NewEncoder(w).Encode(LemonSqueezyResponse{
				Activated: true,
				License: struct {
					Status string `json:"status"`
				}{Status: "active"},
			})
		} else {
			// Simulate Failure
			_ = json.NewEncoder(w).Encode(LemonSqueezyResponse{
				Activated: false,
				Error:     "License key not found",
			})
		}
	}))
	defer mockServer.Close()

	// Setup Manager
	tmpDir, _ := os.MkdirTemp("", "license_api_test")
	defer func() { _ = os.RemoveAll(tmpDir) }()
	manager := NewLicenseManager(tmpDir)
	manager.apiURL = mockServer.URL + "/v1/licenses/activate" // Inject Mock URL

	t.Run("Valid Key Activates", func(t *testing.T) {
		success, err := manager.ActivateLicense("VALID-KEY")
		assert.NoError(t, err)
		assert.True(t, success)

		// Verify it saved
		status, _ := manager.CheckStatus()
		assert.Equal(t, StatusActive, status.Status)
	})

	t.Run("Invalid Key Fails", func(t *testing.T) {
		success, err := manager.ActivateLicense("INVALID-KEY")
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "License key not found")
	})
}
