package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/licensing"
)

func (ui *UI) makeLoginUI(loginWindow fyne.Window) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Nombre de usuario")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Contraseña")

	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Nombre de Usuario", Widget: usernameEntry},
			{Text: "Contraseña", Widget: passwordEntry},
		},
		SubmitText: "Ingresar",
		OnSubmit: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			user, err := ui.Services.UserService.Login(ctx, usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				dialog.ShowError(err, loginWindow)
				return
			}
			ui.currentUser = user

			// Open main window first
			ui.openMainWindow()

			// Hide login window immediately to give visual feedback
			loginWindow.Hide()

			// Close it slightly later to ensure events are finished
			go func() {
				time.Sleep(100 * time.Millisecond)
				loginWindow.Close()
			}()
		},
	}

	forgotPassBtn := widget.NewButton("¿Olvidaste tu contraseña?", func() {
		ui.showRecoveryDialog(loginWindow)
	})
	forgotPassBtn.Importance = widget.LowImportance

	// Use a grid with spacers to center the form and give it a reasonable width
	grid := container.NewGridWithRows(3,
		layout.NewSpacer(),
		container.NewVBox(
			container.NewPadded(loginForm),
			container.NewCenter(forgotPassBtn),
		),
		layout.NewSpacer(),
	)

	return grid
}

func (ui *UI) showRecoveryDialog(parent fyne.Window) {
	licenseEntry := widget.NewEntry()
	licenseEntry.SetPlaceHolder("Ingresa tu clave de licencia")

	dialog.ShowCustomConfirm("Recuperación de Acceso", "Verificar", "Cancelar", licenseEntry, func(confirm bool) {
		if !confirm || licenseEntry.Text == "" {
			return
		}

		// Validar Licencia
		configDir, _ := os.UserConfigDir()
		licensePath := filepath.Join(configDir, "Verith")
		licMgr := licensing.NewLicenseManager(licensePath)

		progress := dialog.NewCustomWithoutButtons("Verificando licencia...", widget.NewProgressBarInfinite(), parent)
		progress.Show()

				go func() {

					defer progress.Hide()

					status, err := licMgr.ValidateLicense(licenseEntry.Text, "")

					

					// Si la validación falla o no está activa

		
			if err != nil || status != "active" {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("Licencia inválida o expirada. No se puede autorizar la recuperación."), parent)
				})
				return
			}

			// Licencia válida -> Cargar Admins
			ctx := context.Background()
			admins, err := ui.Services.UserService.GetAdminUsers(ctx)
			if err != nil || len(admins) == 0 {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("No se encontraron administradores para recuperar."), parent)
				})
				return
			}

			// Mostrar diálogo de reset
			fyne.Do(func() {
				ui.showResetPasswordDialog(parent, admins)
			})
		}()
	}, parent)
}

func (ui *UI) showResetPasswordDialog(parent fyne.Window, admins []domain.User) {
	adminNames := make([]string, len(admins))
	for i, u := range admins {
		adminNames[i] = u.Username
	}

	adminSelect := widget.NewSelect(adminNames, nil)
	if len(adminNames) > 0 {
		adminSelect.SetSelected(adminNames[0])
	}

	newPass := widget.NewPasswordEntry()
	confirmPass := widget.NewPasswordEntry()

	form := widget.NewForm(
		widget.NewFormItem("Usuario a recuperar", adminSelect),
		widget.NewFormItem("Nueva Contraseña", newPass),
		widget.NewFormItem("Confirmar Contraseña", confirmPass),
	)

	dialog.ShowCustomConfirm("Restablecer Contraseña", "Guardar", "Cancelar", form, func(confirm bool) {
		if !confirm {
			return
		}
		if newPass.Text != confirmPass.Text {
			dialog.ShowError(fmt.Errorf("las contraseñas no coinciden"), parent)
			return
		}
		if len(newPass.Text) < 8 {
			dialog.ShowError(fmt.Errorf("la contraseña debe tener al menos 8 caracteres"), parent)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := ui.Services.UserService.ResetPassword(ctx, adminSelect.Selected, newPass.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error al actualizar: %v", err), parent)
			return
		}

		dialog.ShowInformation("Éxito", "Contraseña actualizada. Ahora puedes iniciar sesión.", parent)
	}, parent)
}
