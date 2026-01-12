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

	// Use a grid with spacers to center the form and give it a reasonable width
	grid := container.NewGridWithRows(3,
		layout.NewSpacer(),
		container.NewPadded(loginForm),
		layout.NewSpacer(),
	)

	return grid
}
