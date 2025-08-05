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
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			user, err := ui.Services.UserService.Login(ctx, usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				dialog.ShowError(err, ui.mainWindow)
				return
			}

			ui.currentUser = user
			ui.buildMainUI()
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
