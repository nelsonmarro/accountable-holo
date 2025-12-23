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

			// 2. Build the UI in a background goroutine to avoid blocking the main thread
			go func() {
				// Give the loading screen a moment to render
				time.Sleep(300 * time.Millisecond)

				// Create the base tab container
				tabs := container.NewAppTabs()

				// Update the window to show the empty tabs and resize (still lightweight)
				fyne.Do(func() {
					ui.mainWindow.SetContent(tabs)
					ui.mainWindow.SetMainMenu(ui.makeMainMenu())
					ui.mainWindow.Resize(fyne.NewSize(1280, 720))
					ui.mainWindow.CenterOnScreen()
					ui.mainWindow.SetFullScreen(true) // Enter fullscreen smoothly
				})

				// 3. Build and add tabs one by one
				// This breaks the heavy lifting into smaller chunks for the renderer

				// Summary Tab
				summaryTabContent := ui.makeSummaryTab()
				fyne.Do(func() {
					tabs.Append(container.NewTabItemWithIcon("Resumen Financiero", theme.HomeIcon(), summaryTabContent))
					// Load data for the first visible tab
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

				// 4. Finally, attach the lazy loading logic
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
