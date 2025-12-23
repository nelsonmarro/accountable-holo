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

			// 1. Show a lightweight loading screen immediately
			// This keeps the UI responsive while we prepare the main view.
			loadingContent := container.NewCenter(
				container.NewVBox(
					widget.NewLabelWithStyle("Cargando...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
					widget.NewProgressBarInfinite(),
				),
			)
			ui.mainWindow.SetContent(loadingContent)

			// 2. Build the UI in a background goroutine
			go func() {
				// Give the loading screen a moment to render
				time.Sleep(300 * time.Millisecond)

				// Create the base tab container
				tabs := container.NewAppTabs()

				// Build ONLY the Summary Tab initially (it's the first view)
				summaryTabContent := ui.makeSummaryTab()

				// Create placeholders for other tabs to be loaded later
				accountsTabContent := widget.NewLabel("Cargando Cuentas...")
				txTabContent := widget.NewLabel("Cargando Transacciones...")

				var userTabContent fyne.CanvasObject
				if ui.currentUser.Role == domain.AdminRole {
					userTabContent = widget.NewLabel("Cargando Usuarios...")
				}

				// Update the window content on the main thread
				fyne.Do(func() {
					tabs.Append(container.NewTabItemWithIcon("Resumen Financiero", theme.HomeIcon(), summaryTabContent))
					tabs.Append(container.NewTabItemWithIcon("Cuentas", theme.StorageIcon(), accountsTabContent))
					tabs.Append(container.NewTabItemWithIcon("Transacciones", theme.ListIcon(), txTabContent))

					if userTabContent != nil {
						tabs.Append(container.NewTabItemWithIcon("Usuarios", theme.AccountIcon(), userTabContent))
					}

					ui.mainWindow.SetContent(tabs)
					ui.mainWindow.SetMainMenu(ui.makeMainMenu())
					ui.mainWindow.Resize(fyne.NewSize(1280, 720))
					ui.mainWindow.CenterOnScreen()

					// Setup the lazy loading logic to replace placeholders with real content
					lazyLoadDbCalls(tabs, ui)

					// Load data for the initial tab
					go ui.loadAccountsForSummary()
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
