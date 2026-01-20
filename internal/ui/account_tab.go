package ui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/ui/componets/account"
)

func (ui *UI) makeAccountTab() fyne.CanvasObject {
	// Title
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Cuentas",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	// List
	ui.accountList = widget.NewList(
		func() int {
			return len(ui.accounts)
		},
		ui.makeAccountListUI,
		ui.fillAccountListData,
	)

	// Add Account Button
	accountAddBtn := widget.NewButtonWithIcon("Agregar Cuenta", theme.ContentAddIcon(), func() {
		dialogHandler := account.NewAddAccountDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.AccService,
			ui.loadAccounts,
		)
		dialogHandler.Show()
	})
	accountAddBtn.Importance = widget.HighImportance

	// Containers
	titleContainer := container.NewVBox(
		container.NewCenter(title),
		container.NewBorder(nil, nil, accountAddBtn, nil, nil),
	)

	tableHeader := container.NewGridWithColumns(5,
		widget.NewLabel("Nombre"),
		widget.NewLabel("Número"),
		widget.NewLabel("Tipo"),
		widget.NewLabel("Balance Inicial"),
		widget.NewLabel("Acciones"),
	)

	tableContainer := container.NewBorder(
		container.NewPadded(tableHeader), nil, nil, nil,
		ui.accountList,
	)

	mainContent := container.NewBorder(
		container.NewPadded(titleContainer),
		nil, nil, nil,
		tableContainer,
	)

	return mainContent
}

func (ui *UI) loadAccounts() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	accounts, err := ui.Services.AccService.GetAllAccounts(ctx)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.accounts = accounts

	fyne.Do(func() {
		ui.accountList.Refresh()
	})
}

func (ui *UI) makeAccountListUI() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Importance = widget.DangerImportance

	return container.NewGridWithColumns(5,
		widget.NewLabel("template name"),
		widget.NewLabel("template number"),
		widget.NewLabel("template type"),
		widget.NewLabel("template balance"),
		container.NewHBox(
			editBtn,
			deleteBtn,
		),
	)
}

func (ui *UI) fillAccountListData(i widget.ListItemID, o fyne.CanvasObject) {
	acc := ui.accounts[i]
	rowContainer := o.(*fyne.Container)

	rowContainer.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s - %s", acc.Name, acc.Number))
	rowContainer.Objects[1].(*widget.Label).SetText(acc.Number)
	rowContainer.Objects[2].(*widget.Label).SetText(fmt.Sprintf("Tipo de Cuenta: %s", string(acc.Type)))
	rowContainer.Objects[3].(*widget.Label).SetText(strconv.FormatFloat(acc.InitialBalance, 'f', 2, 64))

	actionsContainer := rowContainer.Objects[4].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	deleteBtn := actionsContainer.Objects[1].(*widget.Button)

	editBtn.OnTapped = func() {
		dialogHandler := account.NewEditAccountDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.AccService,
			ui.loadAccounts,
			acc.ID,
		)
		dialogHandler.Show()
	}

	deleteBtn.OnTapped = func() {
		dialogHandler := account.NewDeleteAccountDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.AccService,
			ui.loadAccounts,
			acc.ID,
		)
		dialogHandler.Show()
	}
}
