package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeUserTab() fyne.CanvasObject {
	// Title
	title := widget.NewLabel("Manage Users")

	// User List
	userList := widget.NewList(
		func() int {
			// TODO: Replace with actual user data
			return 0
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// TODO: Update with actual user data
		},
	)

	// Add User Button
	addUserBtn := widget.NewButtonWithIcon("Add User", theme.ContentAddIcon(), func() {
		// TODO: Implement Add User Dialog
	})

	// Toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() { addUserBtn.OnTapped() }),
	)

	// Layout
	content := container.NewBorder(
		container.NewVBox(title, toolbar),
		nil, nil, nil,
		userList,
	)

	return content
}
