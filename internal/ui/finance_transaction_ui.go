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
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets/transaction"
)

func (ui *UI) makeTransactionUI() fyne.CanvasObject {
	// Title
	title := widget.NewRichText(&widget.TextSegment{
		Text: "Transacciones",
		Style: widget.RichTextStyle{
			SizeName:  theme.SizeNameHeadingText,
			Alignment: fyne.TextAlignCenter,
		},
	})

	// Search Bar
	searchBar := componets.NewSearchBar(ui.filterTransactions)

	// Pagination and List
	ui.transactionPaginator = componets.NewPagination(
		func() (totalCount int) {
			if ui.transactions == nil {
				return 0
			}
			return int(ui.transactions.TotalCount)
		},
		ui.loadTransactions,
		pageSizeOpts...,
	)

	ui.transactionList = widget.NewList(
		func() int {
			if ui.transactions == nil {
				return 0
			}
			return len(ui.transactions.Data)
		},
		ui.makeTransactionListUI,
		ui.fillTransactionListData,
	)

	// Account selector
	ui.accountSelector = widget.NewSelect(
		[]string{},
		func(s string) {
			for _, acc := range ui.accounts {
				if acc.Name == s {
					ui.selectedAccountID = acc.ID
					go ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
					break
				}
			}
		},
	)
	go ui.loadAccountsForTx()

	// Add Transaction Button
	txAddBtn := widget.NewButtonWithIcon("Agregar Transacción", theme.ContentAddIcon(), func() {
		dialogHanlder := transaction.NewAddTransactionDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.TxService,
			ui.Services.CatService,
			func() {
				ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
			},
			ui.selectedAccountID)

		dialogHanlder.Show()
	})
	txAddBtn.Importance = widget.HighImportance

	// Containers
	topBar := container.NewBorder(nil, nil, txAddBtn, nil, searchBar)
	filters := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Cuenta:"),
		nil,
		ui.accountSelector,
	)

	titleContainer := container.NewVBox(
		container.NewCenter(title),
		container.NewPadded(filters),
		topBar,
	)

	tableHeader := container.NewBorder(
		ui.transactionPaginator,
		nil, nil, nil,
		container.NewGridWithColumns(8,
			widget.NewLabel("#"),
			widget.NewLabel("Fecha"),
			widget.NewLabel("Descripción"),
			widget.NewLabel("Categoría"),
			widget.NewLabel("Tipo"),
			widget.NewLabel("Monto"),
			widget.NewLabel("Saldo"),
			widget.NewLabel("Acciones"),
		),
	)

	tableContainer := container.NewBorder(
		tableHeader, nil, nil, nil,
		container.NewHScroll(ui.transactionList),
	)

	mainContent := container.NewBorder(
		container.NewPadded(titleContainer),
		nil, nil, nil,
		tableContainer,
	)

	return mainContent
}

func (ui *UI) makeTransactionListUI() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	voidBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
	voidBtn.Importance = widget.DangerImportance

	lblTxNumber := widget.NewLabel("template number")
	lblTxNumber.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblDate := widget.NewLabel("2025-07-03")
	lblDate.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblDescription := widget.NewLabel("template description")
	lblDescription.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblCategory := widget.NewLabel("template category")
	lblCategory.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblType := widget.NewLabel("template type")
	lblType.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblAmount := widget.NewLabel("$1,200.00")
	lblAmount.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	lblBalance := widget.NewLabel("$5,250.50")
	lblBalance.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	grid := container.NewGridWithColumns(8,
		lblTxNumber,
		lblDate,
		lblDescription,
		lblCategory,
		lblType,
		lblAmount,
		lblBalance,
		container.NewHBox(
			editBtn,
			voidBtn,
		),
	)
	return grid
}

func (ui *UI) fillTransactionListData(i widget.ListItemID, o fyne.CanvasObject) {
	tx := ui.transactions.Data[i]

	rowContainer := o.(*fyne.Container)

	tIDLabel := rowContainer.Objects[0].(*widget.Label)
	tIDLabel.SetText(tx.TransactionNumber)

	dateLabel := rowContainer.Objects[1].(*widget.Label)
	dateLabel.SetText(tx.TransactionDate.Format("2006-01-02"))

	descLabel := rowContainer.Objects[2].(*widget.Label)
	descLabel.SetText(tx.Description)

	categoryLabel := rowContainer.Objects[3].(*widget.Label)
	if tx.Category != nil {
		categoryLabel.SetText(tx.Category.Name)
	} else {
		categoryLabel.SetText("-")
	}

	typeLabel := rowContainer.Objects[4].(*widget.Label)
	if tx.Category != nil {
		typeLabel.SetText(string(tx.Category.Type))
	} else {
		typeLabel.SetText("-")
	}

	amountLabel := rowContainer.Objects[5].(*widget.Label)
	amountText := fmt.Sprintf("%.2f", tx.Amount)
	if tx.Category != nil && tx.Category.Type == domain.Income {
		amountText = "+ $" + amountText
	} else {
		amountText = "- $" + amountText
	}
	amountLabel.SetText(amountText)

	balanceLabel := rowContainer.Objects[6].(*widget.Label)
	balanceLabel.SetText(fmt.Sprintf("$%.2f", tx.RunningBalance))

	actionsContainer := rowContainer.Objects[7].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	voidBtn := actionsContainer.Objects[1].(*widget.Button)

	editBtn.OnTapped = func() {
		dialigHandler := transaction.NewEditTransactionDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.TxService,
			ui.Services.CatService,
			func() {
				ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
			},
			tx.ID,
			ui.selectedAccountID,
		)

		dialigHandler.Show()
	}
	voidBtn.OnTapped = func() {
		dialog.ShowInformation("Info", "Void transaction not implemented yet.", ui.mainWindow)
	}
}

func (ui *UI) loadTransactions(page int, pageSize int) {
	if ui.selectedAccountID == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := ui.Services.TxService.GetTransactionByAccountPaginated(ctx, ui.selectedAccountID, page, pageSize, ui.transactionFilter)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.transactions = result

	fyne.Do(func() {
		ui.transactionList.Refresh()
		ui.transactionPaginator.Refresh()
	})
}

func (ui *UI) filterTransactions(filter string) {
	ui.transactionFilter = filter
	ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
}

func (ui *UI) loadAccountsForTx() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Assuming page 1 and a large enough page size to get all accounts
	result, err := ui.Services.AccService.GetAllAccounts(ctx)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.accounts = result
	accountNames := make([]string, len(ui.accounts))
	for i, acc := range ui.accounts {
		accountNames[i] = acc.Name
	}

	fyne.Do(func() {
		ui.accountSelector.Options = accountNames
		if len(ui.accounts) > 0 {
			ui.selectedAccountID = ui.accounts[0].ID
			ui.accountSelector.SetSelected(ui.accounts[0].Name)
		}
		ui.accountSelector.Refresh()
	})
	go ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
}
