package ui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

func (ui *UI) makeLoginUI() fyne.CanvasObject {
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
				dialog.ShowError(err, ui.mainWindow)
				return
			}
			ui.currentUser = user

			// 1. Create and set the empty tab container immediately
			tabs := ui.buildMainUI()
			ui.mainWindow.SetContent(tabs)
			ui.mainWindow.Resize(fyne.NewSize(1280, 720))
			ui.mainWindow.CenterOnScreen()

			// 2. Build and append each tab asynchronously
			go func() {
				// Summary Tab
				summaryTabContent := ui.makeSummaryTab()
				fyne.Do(func() {
					tabs.Append(container.NewTabItemWithIcon("Resumen Financiero", theme.HomeIcon(), summaryTabContent))
					// Since this is the first tab, load its data
					go ui.loadAccountsForSummary()
				})

				// Accounts Tab
				accountsTabContent := ui.makeAccountTab()
				fyne.Do(func() {
					tabs.Append(container.NewTabItemWithIcon("Cuentas", theme.StorageIcon(), accountsTabContent))
				})

				// Transactions Tab
				txTabContent := ui.makeFinancesTab()
				fyne.Do(func() {
					tabs.Append(container.NewTabItemWithIcon("Transacciones", theme.ListIcon(), txTabContent))
				})

				// User Tab (conditional)
				if ui.currentUser.Role == domain.AdminRole {
					userTabContent := ui.makeUserTab()
					fyne.Do(func() {
						tabs.Append(container.NewTabItemWithIcon("Usuarios", theme.AccountIcon(), userTabContent))
					})
				}

				// Finally, set up the lazy loading for subsequent clicks
				fyne.Do(func() {
					lazyLoadDbCalls(tabs, ui)
				})
			}()
		},
	}

	// Use a grid with spacers to center the form and give it a reasonable width
	grid := container.NewGridWithRows(3,
		layout.NewSpacer(),
		container.NewPadded(loginForm),
		layout.NewSpacer(),
	)

	return grid
}
