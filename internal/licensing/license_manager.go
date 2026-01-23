package licensing

import (
	"time"

	"github.com/zalando/go-keyring"
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

	// --- Seguridad: Validación Cruzada con Keyring ---
	// Intentamos recuperar la fecha de instalación "real" del sistema seguro
	const serviceName = "Verith"
	const userKey = "InstallDate"
	
	storedDateStr, errKeyring := keyring.Get(serviceName, userKey)
	
	if errKeyring == nil {
		// Encontramos una fecha en el llavero. Parsearla.
		realInstallDate, errParse := time.Parse(time.RFC3339, storedDateStr)
		if errParse == nil {
			// Si la fecha del archivo es POSTERIOR a la del llavero, el usuario borró el archivo.
			// Restauramos la fecha original.
			if data.InstallDate.After(realInstallDate) {
				data.InstallDate = realInstallDate
				// Forzamos guardar el archivo corregido para que el usuario vea la realidad
				_ = m.store.Save(data)
			}
		}
	} else {
		// No existe en el llavero (primera vez real o limpieza total).
		// Guardamos la fecha actual del archivo en el llavero como ancla.
		_ = keyring.Set(serviceName, userKey, data.InstallDate.Format(time.RFC3339))
	}
	// ------------------------------------------------

	// Si ya esta activo, verificamos la suscripción periódicamente
	if data.Status == StatusActive {
		// Validar online cada 24 horas para asegurar que la suscripción sigue pagada
		if time.Since(data.LastCheck).Hours() > 24 {
			status, err := m.ValidateLicense(data.LicenseKey, data.InstanceID)
			if err == nil {
				// Conexión exitosa, actualizamos estado real
				if status == "active" {
					data.LastCheck = time.Now()
					_ = m.store.Save(data)
				} else {
					// Suscripción vencida o cancelada
					data.Status = StatusExpired
					_ = m.store.Save(data)
				}
			}
			// Si err != nil (Error de conexión): Permitimos el acceso (Grace Period)
			// No actualizamos LastCheck para reintentar la próxima vez
		}
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
