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
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Manage Users",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	// User List
	ui.userList = widget.NewList(
		func() int {
			return len(ui.users)
		},
		ui.createUserItem,
		ui.updateUserItem,
	)

	// Add User Button
	addUserBtn := widget.NewButtonWithIcon("Add User", theme.ContentAddIcon(), func() {
		dialogHandler := user.NewAddUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, ui.currentUser)
		dialogHandler.Show()
	})
	addUserBtn.Importance = widget.HighImportance

	// Reload Data Button
	reloadDataBtn := widget.NewButtonWithIcon("Reload Data", theme.ViewRefreshIcon(), func() {
		ui.loadUsers()
	})
	reloadDataBtn.Importance = widget.SuccessImportance

	// Toolbar
	topBar := container.NewBorder(nil, nil, container.NewHBox(addUserBtn), container.NewHBox(reloadDataBtn))

	titleContainer := container.NewVBox(
		container.NewCenter(title),
		topBar,
	)

	tableHeader := container.NewGridWithColumns(3,
		widget.NewLabel("Username"),
		widget.NewLabel("Role"),
		widget.NewLabel("Actions"),
	)

	tableContainer := container.NewBorder(
		tableHeader, nil, nil, nil,
		ui.userList,
	)

	mainContent := container.NewBorder(
		container.NewPadded(titleContainer),
		nil, nil, nil,
		tableContainer,
	)

	go ui.loadUsers()

	return mainContent
}

func (ui *UI) createUserItem() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Importance = widget.DangerImportance

	return container.NewGridWithColumns(3,
		widget.NewLabel("username template"),
		widget.NewLabel("role template"),
		container.NewHBox(editBtn, deleteBtn),
	)
}

func (ui *UI) updateUserItem(i widget.ListItemID, o fyne.CanvasObject) {
	userToUpdate := ui.users[i]
	grid := o.(*fyne.Container)

	usernameLabel := grid.Objects[0].(*widget.Label)
	usernameLabel.SetText(userToUpdate.Username)

	roleLabel := grid.Objects[1].(*widget.Label)
	roleLabel.SetText(string(userToUpdate.Role))

	actionsContainer := grid.Objects[2].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	editBtn.OnTapped = func() {
		dialogHandler := user.NewEditUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, ui.currentUser, &userToUpdate)
		dialogHandler.Show()
	}

	deleteBtn := actionsContainer.Objects[1].(*widget.Button)
	deleteBtn.OnTapped = func() {
		dialogHandler := user.NewDeleteUserDialog(ui.mainWindow, ui.errorLogger, ui.Services.UserService, func() { ui.loadUsers() }, &userToUpdate, ui.currentUser)
		dialogHandler.Show()
	}
}

func (ui *UI) loadUsers() {
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
		ui.userList.Refresh()
	})
}
