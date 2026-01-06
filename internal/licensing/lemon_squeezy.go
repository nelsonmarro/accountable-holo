package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	// Datos a enviar
	// Lemon squeezy espera el lincense_key y opcionalmente un instance_name
	payload := map[string]string{
		"license_key":   key,
		"instance_name": "AccountableHolo",
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
