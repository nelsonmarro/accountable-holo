package ui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/ui/components/account"
)

func (ui *UI) makeAccountTab() fyne.CanvasObject {
	// Title
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Cuentas",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText, // Use the heading size from our custom theme
			Alignment: fyne.TextAlignCenter,
		},
	})

	// Pagination and List
	ui.accountList = widget.NewList(
		func() int {
			return len(ui.accounts)
		}, ui.makeListUI, ui.fillListData,
	)
	go ui.loadAccounts()

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
	headerArea := container.NewVBox(
		container.NewCenter(title),
		container.NewHBox(layout.NewSpacer(), accountAddBtn),
	)
	mainContent := container.NewBorder(container.NewPadded(headerArea), nil, nil, nil, ui.accountList)

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

func (ui *UI) makeListUI() fyne.CanvasObject {
	transactionBtn := widget.NewButtonWithIcon("Transacciones", theme.ListIcon(), func() {})

	editBtn := widget.NewButtonWithIcon("Editar", theme.DocumentCreateIcon(), func() {})
	editBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("Eliminar", theme.DeleteIcon(), func() {})
	deleteBtn.Importance = widget.DangerImportance

	return container.NewBorder(
		nil,
		nil,
		container.NewVBox(
			container.NewHBox(
				widget.NewLabel("nombre de cuenta"),
				widget.NewLabel("tipo de cuenta"),
			),
			container.NewHBox(
				widget.NewLabel("Balance:"),
				widget.NewLabel("balance inicial"),
			),
		),
		container.NewPadded(
			container.NewHBox(
				transactionBtn,
				editBtn,
				deleteBtn,
			),
		),
	)
}

func (ui *UI) fillListData(i widget.ListItemID, o fyne.CanvasObject) {
	borderContainer := o.(*fyne.Container)
	infoContainer := borderContainer.Objects[0].(*fyne.Container)
	paddedContainer := borderContainer.Objects[1].(*fyne.Container)
	buttonsContainer := paddedContainer.Objects[0].(*fyne.Container)

	cuentaInfoContainer := infoContainer.Objects[0].(*fyne.Container)
	nameLbl := cuentaInfoContainer.Objects[0].(*widget.Label)
	nameLbl.SetText(fmt.Sprintf("%s - %s", ui.accounts[i].Name, ui.accounts[i].Number))

	typeLbl := cuentaInfoContainer.Objects[1].(*widget.Label)
	typeLbl.SetText(fmt.Sprintf("Tipo de Cuenta: %s", string(ui.accounts[i].Type)))

	cuentaBalanceContainer := infoContainer.Objects[1].(*fyne.Container)
	balanceLbl := cuentaBalanceContainer.Objects[1].(*widget.Label)
	balanceText := strconv.FormatFloat(ui.accounts[i].InitialBalance, 'f', 2, 64)
	balanceLbl.SetText(balanceText)

	deleteBtn := buttonsContainer.Objects[2].(*widget.Button)
	deleteBtn.OnTapped = func() {
		// Create an instance of the dialog handler and show it
		dialogHandler := account.NewDeleteAccountDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.AccService,
			ui.loadAccounts,
			ui.accounts[i].ID, // Pass the specific ID for this row
		)
		dialogHandler.Show()
	}

	editBtn := buttonsContainer.Objects[1].(*widget.Button)
	editBtn.OnTapped = func() {
		dialogHandler := account.NewEditAccountDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.AccService,
			ui.loadAccounts,
			ui.accounts[i].ID, // Pass the specific ID for this row
		)
		dialogHandler.Show()
	}
}
