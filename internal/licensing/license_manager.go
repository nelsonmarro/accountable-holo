package licensing

import (
	"time"
)

type LicenseManager struct {
	store  *LicenseStore
	apiURL string // Optional: For testing
}

func NewLicenseManager(appDataPath string) *LicenseManager {
	return &LicenseManager{
		store: NewLicenseStore(appDataPath),
	}
}

// CheckStatus verifica el estado actual y actualiza si el trial expiró
func (m *LicenseManager) CheckStatus() (*LicenseData, error) {
	data, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	// Si ya esta activo, no hacemos nada
	if data.Status == StatusActive {
		return data, nil
	}

	// Logica de Trial: ¿Pasaron más de 15 días desde la instalación?
	daysSinceInstall := int(time.Since(data.InstallDate).Hours() / 24)

	if daysSinceInstall > 15 {
		data.Status = StatusExpired
		// Guardamos el cambio de estado a expirado
		err := m.store.Save(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
