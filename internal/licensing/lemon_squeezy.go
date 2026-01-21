package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type LemonSqueezyResponse struct {
	Activated bool   `json:"activated"`
	Error     string `json:"error,omitempty"`
	License   struct {
		Status string `json:"status"`
	} `json:"license"`
}

// ActivateLicense llama a la API de Lemon Squeezy
func (m *LicenseManager) ActivateLicense(key string) (bool, error) {
	// Endpoint oficial de activación de Lemon Squeezy
	url := "https://api.lemonsqueezy.com/v1/licenses/activate"
	if m.apiURL != "" {
		url = m.apiURL
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Verith-PC"
	}

	// Datos a enviar
	// Usamos el hostname para identificar cada puesto de trabajo único
	payload := map[string]string{
		"license_key":   key,
		"instance_name": hostname,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	// Crear la solicitud http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var lsResp LemonSqueezyResponse
	if err := json.NewDecoder(resp.Body).Decode(&lsResp); err != nil {
		return false, fmt.Errorf("error al leer respuesta del servidor")
	}

	// Error devuelto por la API
	if lsResp.Error != "" {
		return false, fmt.Errorf("%s", lsResp.Error)
	}

	// Si se activo correctamente
	if lsResp.Activated {
		// Actualizamos en el estado la licencia
		data, _ := m.store.Load()
		data.Status = StatusActive
		data.LicenseKey = key
		data.LastCheck = time.Now()
		_ = m.store.Save(data)

		return true, nil
	}

	return false, fmt.Errorf("la licencia no pudo ser activada")
}

// ValidateLicense verifica el estado actual de la suscripción sin consumir activaciones
func (m *LicenseManager) ValidateLicense(key string) (string, error) {
	url := "https://api.lemonsqueezy.com/v1/licenses/validate"
	if m.apiURL != "" {
		url = m.apiURL
	}
	// Use instance_id if available, but validate mainly needs license_key
	payload := map[string]string{
		"license_key": key,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	var lsResp struct {
		Valid   bool   `json:"valid"`
		Error   string `json:"error"`
		LicenseKey struct {
			Status string `json:"status"` // "active", "expired", "inactive"
		} `json:"license_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&lsResp); err != nil {
		return "", fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if lsResp.Error != "" {
		return "", fmt.Errorf("api error: %s", lsResp.Error)
	}

	// Retorna el estado real de la suscripción (active, expired, etc.)
	return lsResp.LicenseKey.Status, nil
}