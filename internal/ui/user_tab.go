package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/user"
)

func (ui *UI) makeUserTab() fyne.CanvasObject {
	// Title
	title := widget.NewLabel("Manage Users")

	// User List
	userList := widget.NewList(
		func() int {
			return len(ui.users)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("", theme.DeleteIcon(), nil), widget.NewLabel("template"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			label.SetText(ui.users[id].Username)

			editBtn := container.Objects[1].(*widget.Button)
			editBtn.OnTapped = func() {
				dialogHandler := user.NewEditUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, ui.currentUser, &ui.users[id])
				dialogHandler.Show()
			}

			deleteBtn := container.Objects[2].(*widget.Button)
			deleteBtn.OnTapped = func() {
				dialogHandler := user.NewDeleteUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, &ui.users[id], ui.currentUser)
				dialogHandler.Show()
			}
		},
	)

	// Add User Button
	addUserBtn := widget.NewButtonWithIcon("Add User", theme.ContentAddIcon(), func() {
		dialogHandler := user.NewAddUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, ui.currentUser)
		dialogHandler.Show()
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

	ui.loadUsers = func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			users, err := ui.Services.UserService.GetAllUsers(ctx, ui.currentUser)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("error loading users: %w", err), ui.mainWindow)
				})
				return
			}

			fyne.Do(func() {
				ui.users = users
				userList.Refresh()
			})
		}()
	}

	ui.loadUsers()

	return content
}
