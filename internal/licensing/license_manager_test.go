package licensing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseManager_CheckStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "verith-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("New Installation - Trial Status", func(t *testing.T) {
		mgr := NewLicenseManager(tmpDir)
		data, err := mgr.CheckStatus()
		
		assert.NoError(t, err)
		assert.Equal(t, StatusTrial, data.Status)
		assert.NotEmpty(t, data.InstallDate)
	})

	t.Run("Trial Expired after 15 days", func(t *testing.T) {
		mgr := NewLicenseManager(tmpDir)
		
		// Manipular el archivo para que parezca antiguo
		data, _ := mgr.store.Load()
		data.InstallDate = time.Now().AddDate(0, 0, -16)
		data.Status = StatusTrial
		_ = mgr.store.Save(data)

		status, err := mgr.CheckStatus()
		assert.NoError(t, err)
		assert.Equal(t, StatusExpired, status.Status)
	})

	t.Run("Active License - No Online Check if < 24h", func(t *testing.T) {
		mgr := NewLicenseManager(tmpDir)
		
		data, _ := mgr.store.Load()
		data.Status = StatusActive
		data.LastCheck = time.Now().Add(-1 * time.Hour)
		_ = mgr.store.Save(data)

		status, err := mgr.CheckStatus()
		assert.NoError(t, err)
		assert.Equal(t, StatusActive, status.Status)
	})
}

func TestLicenseManager_OnlineValidation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "verith-online-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Mock Server para Lemon Squeezy
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Respuesta simulada de validación exitosa
		w.Write([]byte(`{
			"valid": true,
			"license_key": {
				"status": "active"
			}
		}`))
	}))
	defer ts.Close()

	t.Run("CheckStatus triggers Online Validation after 24h", func(t *testing.T) {
		mgr := NewLicenseManager(tmpDir)
		mgr.apiURL = ts.URL // Inyectar URL del mock

		data, _ := mgr.store.Load()
		data.Status = StatusActive
		data.LicenseKey = "TEST-KEY"
		data.LastCheck = time.Now().Add(-25 * time.Hour) // Forzar chequeo
		_ = mgr.store.Save(data)

		status, err := mgr.CheckStatus()
		assert.NoError(t, err)
		assert.Equal(t, StatusActive, status.Status)
		// Verificar que LastCheck se actualizó a "ahora" (aprox)
		assert.True(t, time.Since(status.LastCheck).Seconds() < 5)
	})
}