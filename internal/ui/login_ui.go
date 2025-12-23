package ui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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

			// 1. Show a loading screen immediately to keep the UI responsive
			loadingContent := container.NewCenter(
				container.NewVBox(
					widget.NewLabelWithStyle("Cargando...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
					widget.NewProgressBarInfinite(),
				),
			)
			ui.mainWindow.SetContent(loadingContent)

			// 2. Perform the heavy lifting (building the main UI) in a goroutine
			go func() {
				// Give the UI a moment to render the loading screen
				time.Sleep(200 * time.Millisecond)

				ui.currentUser = user
				// Build the complex UI tree off the main thread
				newContent := ui.buildMainUI()

				// 3. Switch to the main UI on the main thread
				fyne.Do(func() {
					ui.mainWindow.SetMainMenu(ui.makeMainMenu())
					ui.mainWindow.SetContent(newContent)
					ui.mainWindow.Resize(fyne.NewSize(1280, 720))
					ui.mainWindow.CenterOnScreen()
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
