// Package licensing provides structures and functions for managing software licenses.
package licensing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type LicenseStatus string

const (
	StatusTrial   LicenseStatus = "TRIAL"
	StatusActive  LicenseStatus = "ACTIVE"
	StatusExpired LicenseStatus = "EXPIRED"
)

// LicenseData es lo que guardamos en el disco
type LicenseData struct {
	InstallDate time.Time     `json:"install_date"`
	LicenseKey  string        `json:"license_key"`
	InstanceID  string        `json:"instance_id,omitempty"` // ID de activación único
	Status      LicenseStatus `json:"status"`
	LastCheck   time.Time     `json:"last_check"`
}

// LicenseStore maneja la lectura/escritura del archivo
type LicenseStore struct {
	filePath string
}

func NewLicenseStore(appDataPath string) *LicenseStore {
	return &LicenseStore{filePath: filepath.Join(appDataPath, "license.json")}
}

func (s *LicenseStore) Save(data *LicenseData) error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	// Convertir el struct a JSON (con indentación para legibilidad)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Escribir el archivo
	return os.WriteFile(s.filePath, bytes, 0o600)
}

func (s *LicenseStore) Load() (*LicenseData, error) {
	// Intentar de leer el archivo
	bytes, err := os.ReadFile(s.filePath)
	if err != nil {
		// Si el archivo NO existe, es la primera vez que se abre la app
		if os.IsNotExist(err) {
			newData := &LicenseData{
				InstallDate: time.Now(),
				Status:      StatusTrial,
				LastCheck:   time.Now(),
			}
			if err := s.Save(newData); err != nil {
				return nil, err
			}
			return newData, nil
		}
		return nil, err
	}

	// 2. Si el archivo existe, lo decodificamos
	var data LicenseData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
