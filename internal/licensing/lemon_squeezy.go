package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Estructuras oficiales de respuesta de la API de Licencias de Lemon Squeezy
// Docs: https://docs.lemonsqueezy.com/help/licensing/license-api

type LemonSqueezyLicenseKey struct {
	ID              int       `json:"id"`
	Status          string    `json:"status"` // "active", "inactive", "expired", "disabled"
	Key             string    `json:"key"`
	ActivationLimit int       `json:"activation_limit"`
	ActivationUsage int       `json:"activation_usage"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       *string   `json:"expires_at"` // Null si es de por vida o suscripción activa
}

type LemonSqueezyInstance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LemonSqueezyActivateResponse struct {
	Activated  bool                   `json:"activated"`
	Error      string                 `json:"error,omitempty"`
	LicenseKey LemonSqueezyLicenseKey `json:"license_key"`
	Instance   LemonSqueezyInstance   `json:"instance"`
	Meta       map[string]interface{} `json:"meta"`
}

type LemonSqueezyValidateResponse struct {
	Valid      bool                   `json:"valid"`
	Error      string                 `json:"error,omitempty"`
	LicenseKey LemonSqueezyLicenseKey `json:"license_key"`
	Instance   *LemonSqueezyInstance  `json:"instance,omitempty"` // Puede ser null si no se valida una instancia específica
	Meta       map[string]interface{} `json:"meta"`
}

// ActivateLicense llama a POST v1/licenses/activate
func (m *LicenseManager) ActivateLicense(key string) (bool, error) {
	url := "https://api.lemonsqueezy.com/v1/licenses/activate"
	// Permitir override para tests
	if m.apiURL != "" {
		url = m.apiURL
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Verith-Workstation"
	}

	payload := map[string]string{
		"license_key":   key,
		"instance_name": hostname,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("error construyendo petición: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error de conexión con servidor de licencias: %w", err)
	}
	defer resp.Body.Close()

	// Decodificar respuesta
	var lsResp LemonSqueezyActivateResponse
	if err := json.NewDecoder(resp.Body).Decode(&lsResp); err != nil {
		return false, fmt.Errorf("error interpretando respuesta del servidor")
	}

	// 1. Revisar si hubo error lógico en la API (ej: límite alcanzado, clave no existe)
	if lsResp.Error != "" {
		return false, fmt.Errorf("error de activación: %s", lsResp.Error)
	}

	// 2. Revisar flag de éxito
	if !lsResp.Activated {
		return false, fmt.Errorf("la licencia es válida pero no se pudo activar (posible límite de activaciones)")
	}

	// 3. Guardar estado
	data, _ := m.store.Load()
	data.Status = StatusActive
	data.LicenseKey = lsResp.LicenseKey.Key
	data.InstanceID = lsResp.Instance.ID // Guardamos el ID de instancia
	data.LastCheck = time.Now()

	// Si la API devuelve fecha de expiración, podríamos usarla,
	// pero en suscripciones confiamos en la validación recurrente.

	if err := m.store.Save(data); err != nil {
		return true, fmt.Errorf("licencia activada pero falló al guardar localmente: %w", err)
	}

	return true, nil
}

// ValidateLicense llama a POST v1/licenses/validate
// Verifica si la licencia sigue activa y pagada (si es suscripción)
func (m *LicenseManager) ValidateLicense(key, instanceID string) (string, error) {
	url := "https://api.lemonsqueezy.com/v1/licenses/validate"
	if m.apiURL != "" {
		url = m.apiURL
	}

	payload := map[string]string{
		"license_key": key,
	}
	if instanceID != "" {
		payload["instance_id"] = instanceID
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("offline") // Error de red
	}
	defer resp.Body.Close()

	var lsResp LemonSqueezyValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&lsResp); err != nil {
		return "", fmt.Errorf("error de formato")
	}

	// Si hay un error lógico pero tenemos respuesta, devolvemos el estado si existe, o "inactive"
	if lsResp.Error != "" {
		// Casos como "License key not found" o "License key expired"
		// Lemon suele devolver el objeto license_key aun con error.
		if lsResp.LicenseKey.Status != "" {
			return lsResp.LicenseKey.Status, nil
		}
		return "inactive", nil
	}

	// Retornamos el estado real (active, expired, etc.)
	if lsResp.Valid {
		return lsResp.LicenseKey.Status, nil
	}

	return "inactive", nil
}
