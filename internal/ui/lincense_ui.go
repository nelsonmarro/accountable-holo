package ui

import (
	"fmt"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/licensing"
)

// ShowLincenseWindow muestra la ventana de control de acceso de licencias
func (ui *UI) ShowLincenseWindow(mgr *licensing.LicenseManager, onProceed func()) {
	data, _ := mgr.CheckStatus()

	// Si ya esta activo, procedemos directamente
	if data.Status == licensing.StatusActive {
		onProceed()
		return
	}

	window := ui.app.NewWindow("Licencia - Accountable Holo")

	// Titulo y mensaje según el estado
	title := widget.NewLabelWithStyle("Gestión de Licencia", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	message := widget.NewLabel("")
	message.Wrapping = fyne.TextWrapWord
	message.Alignment = fyne.TextAlignCenter

	proceedBtn := widget.NewButton("Continuar con la Prueba", func() {
		window.Close()
		onProceed()
	})
	proceedBtn.Importance = widget.LowImportance // Botón secundario

	if data.Status == licensing.StatusExpired {
		message.SetText("Tu periodo de prueba de 15 días ha expirado.\nPor favor, adquiere una licencia para continuar usando la aplicación.")
		// Opcional: Cambiar color del mensaje a rojo si fuera canvas.Text
		proceedBtn.Disable()
	} else {
		daysLeft := 15 - int(time.Since(data.InstallDate).Hours()/24)
		message.SetText(fmt.Sprintf("Estás en modo de prueba. Te quedan %d días para evaluar todas las funcionalidades.", daysLeft))
	}

	// Formulario de activación
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("XXXX-XXXX-XXXX-XXXX")

	var activateBtn *widget.Button
	activateBtn = widget.NewButtonWithIcon("Activar Licencia", theme.ConfirmIcon(), func() {
		if keyEntry.Text == "" {
			dialog.ShowInformation("Información", "Por favor ingresa una clave de licencia.", window)
			return
		}
		activateBtn.Disable()
		activateBtn.SetText("Verificando...")

		go func() {
			success, err := mgr.ActivateLicense(keyEntry.Text)

			fyne.Do(func() {
				if success {
					dialog.ShowInformation("¡Activación Exitosa!", "Gracias por tu compra. Disfruta de Accountable Holo.", window)
					// Damos tiempo al usuario para leer el mensaje antes de cerrar
					time.AfterFunc(2*time.Second, func() {
						window.Close()
						onProceed()
					})
				} else {
					dialog.ShowError(fmt.Errorf("error de activación: %s", err.Error()), window)
					activateBtn.Enable()
					activateBtn.SetText("Activar Licencia")
				}
			})
		}()
	})
	activateBtn.Importance = widget.HighImportance // Botón primario (destacado)

	buyLink, _ := url.Parse("https://tu-tienda.lemonsqueezy.com") // Pon aquí tu link real luego
	buyBtn := widget.NewHyperlink("Comprar Licencia", buyLink)
	buyBtn.Alignment = fyne.TextAlignCenter

	// Contenedor principal con más aire y ancho expandido

	content := container.NewVBox(
		title,
		message,
		widget.NewSeparator(),
		widget.NewLabel("Ingresa tu clave de licencia:"),
		keyEntry,
		activateBtn,
		container.NewPadded(container.NewCenter(buyBtn)),
		widget.NewSeparator(),
		proceedBtn,
	)

	// Usamos NewPadded para que el contenido se estire a los bordes con margen

	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(500, 450)) // Más ancho y alto
	window.CenterOnScreen()
	window.Show()
}
