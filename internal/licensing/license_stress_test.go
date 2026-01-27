package licensing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// Helper to reset env
func setupTestEnv(t *testing.T) (string, *LicenseManager) {
	tmpDir, err := os.MkdirTemp("", "stress_test_*")
	require.NoError(t, err)
	
	// Mock Keyring (since real keyring depends on OS and can't be easily mocked in integration without side effects)
	// For this stress test, we will focus on file system and API resilience.
	// Note: Real Keyring operations in tests might fail in CI/Headless environments.
	keyring.MockInit() 

	mgr := NewLicenseManager(tmpDir)
	return tmpDir, mgr
}

func TestLicense_CorruptFile(t *testing.T) {
	tmpDir, mgr := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	// Case 1: Corrupt JSON
	path := filepath.Join(tmpDir, "license.json")
	os.WriteFile(path, []byte("{ invalid json ..."), 0644)

	// Should not panic, should probably return error or default to trial
	data, err := mgr.CheckStatus()
	
	// If file is corrupt, Load() fails. CheckStatus should handle it gracefully?
	// Current implementation returns error if Load fails.
	// It forces the user to probably delete the file or contact support. 
	// Ideally, it might treat it as "No License" -> Start Trial, but safety says "Error".
	assert.Error(t, err, "Should report error on corrupt file")
	assert.Nil(t, data)
}

func TestLicense_APITimeout_GracePeriod(t *testing.T) {
	tmpDir, mgr := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	// Mock Server that sleeps/timeouts
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Simulate slow net
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer ts.Close()
	
	mgr.apiURL = ts.URL

	// Setup valid license that needs check (LastCheck > 24h)
	data := &LicenseData{
		Status:      StatusActive,
		LicenseKey:  "VALID",
		LastCheck:   time.Now().Add(-25 * time.Hour),
		InstallDate: time.Now().Add(-100 * time.Hour),
	}
	mgr.store.Save(data)

	// Act
	// Even if API fails (504), user should NOT be blocked immediately (Grace Period logic)
	// And LastCheck should NOT be updated (so it retries next time)
	status, err := mgr.CheckStatus()

	assert.NoError(t, err)
	assert.Equal(t, StatusActive, status.Status, "Should remain active on network error")
	
	// Verify LastCheck was NOT updated (still old)
	// We allow a small margin for execution time, but it should definitely be old
	assert.True(t, time.Since(status.LastCheck).Hours() > 24, "LastCheck should not update on failure")
}

func TestLicense_API_Returns_Expired(t *testing.T) {
	tmpDir, mgr := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	// Mock Server says EXPIRED
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"valid": false, "error": "expired", "license_key": {"status": "expired"}}`)
	}))
	defer ts.Close()
	mgr.apiURL = ts.URL

	// Setup initially active license
	data := &LicenseData{
		Status:      StatusActive,
		LicenseKey:  "WAS-VALID",
		LastCheck:   time.Now().Add(-48 * time.Hour),
	}
	mgr.store.Save(data)

	// Act
	status, err := mgr.CheckStatus()

	assert.NoError(t, err)
	assert.Equal(t, StatusExpired, status.Status, "Should switch to Expired if API says so")
}

func TestLicense_SystemClockManipulation(t *testing.T) {
	// Detecting clock rewind is hard with just local file time.
	// But our logic relies on "InstallDate".
	// If user sets clock to 1990, time.Since(InstallDate) might be huge or negative.
	
	tmpDir, mgr := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	// Simulate Future Install Date (User set clock forward then back? or messed up)
	futureDate := time.Now().Add(24 * time.Hour)
	data := &LicenseData{
		Status:      StatusTrial,
		InstallDate: futureDate,
	}
	mgr.store.Save(data)

	status, _ := mgr.CheckStatus()
	
	// time.Since(future) is negative.
	// Our logic: daysSince = int(Since.Hours()/24)
	// If negative, daysSince < 15. So Trial continues.
	// This is acceptable behavior (benefit of doubt), but we should ensure it doesn't crash or panic.
	assert.Equal(t, StatusTrial, status.Status)
}

func TestLicense_Keyring_Persistence(t *testing.T) {
	// This tests the anti-tamper logic
	tmpDir, mgr := setupTestEnv(t)
	defer os.RemoveAll(tmpDir)

	// 1. Simulate "Old" install in Keyring (30 days ago)
	oldDate := time.Now().Add(-30 * 24 * time.Hour)
	keyring.Set("Verith", "InstallDate", oldDate.Format(time.RFC3339))

	// 2. Simulate "New" fresh file (User deleted license.json)
	// CheckStatus will generate a new file with Today's date inside Load(), 
	// THEN cross-check with keyring.
	
	// We need to trigger the logic.
	// First call to CheckStatus creates the file.
	// Then it checks keyring.
	
	status, err := mgr.CheckStatus()
	require.NoError(t, err)

	// 3. Expectation: The system detects the keyring date (30 days old) 
	// and forces the status to EXPIRED (since 30 > 15).
	
	// Wait... CheckStatus calculates expiration based on data.InstallDate.
	// The cross-check logic updates data.InstallDate to match keyring.
	// So data.InstallDate should be oldDate.
	
	assert.Equal(t, StatusExpired, status.Status, "Should be expired based on Keyring date")
	assert.WithinDuration(t, oldDate, status.InstallDate, 1*time.Second, "InstallDate should be restored from Keyring")
}
