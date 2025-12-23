package ui

import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/helpers"
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
		ui.createTransactiontItem,
		ui.updateTransactionItem,
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
			ui.selectedAccountID,
			*ui.currentUser,
		)

		dialogHanlder.Show()
	})
	txAddBtn.Importance = widget.HighImportance

	// Filtres Buttom
	advancedFiltersBtn := widget.NewButtonWithIcon("Filtros Avanzados", theme.MoreVerticalIcon(), func() {
		filtersDialog := transaction.NewFiltersDialog(
			ui.mainWindow,
			ui.Services.CatService,
			func(filters domain.TransactionFilters,
			) {
				ui.currentTransactionFilters = filters
				ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
			},
			ui.errorLogger,
		)

		filtersDialog.Show()
	})
	advancedFiltersBtn.Importance = widget.HighImportance

	// Report Buttom
	generateReportBtn := widget.NewButtonWithIcon("Reportes", theme.DocumentPrintIcon(), func() {
		reportDialog := componets.NewReportDialog(
			ui.mainWindow,
			func(format string, outputPath string) {
				go ui.generateReportFile(format, outputPath)
			},
			func(format string, outputPath string) {
				go ui.generateDailyReportFile(format, outputPath)
			},
		)
		reportDialog.Show()
	})
	generateReportBtn.Importance = widget.SuccessImportance

	reconciliationBtn := widget.NewButtonWithIcon("Reconciliar", theme.ContentPasteIcon(), func() {
		dialogHandler := transaction.NewReconciliationDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.TxService,
			ui.Services.CatService,
			ui.Services.ReportService,
			ui.accounts,
			func() {
				go ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
			},
			ui.currentUser,
		)

		dialogHandler.Show()
	})
	reconciliationBtn.Importance = widget.WarningImportance

	// Reload Data Button
	reloadDataBtn := widget.NewButtonWithIcon("Recargar Datos", theme.ViewRefreshIcon(), func() {
		ui.currentTransactionFilters = domain.TransactionFilters{} // Reset filters
		ui.transactionSearchText = ""                              // Reset search text
		searchBar.SetText("")
		go ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
	})
	reloadDataBtn.Importance = widget.SuccessImportance

	// Containers
	topBar := container.NewBorder(nil, nil,
		container.NewHBox(txAddBtn, reloadDataBtn),
		container.NewHBox(advancedFiltersBtn, generateReportBtn, reconciliationBtn),
		searchBar,
	)
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
		container.NewGridWithColumns(9,
			widget.NewLabel("#"),
			widget.NewLabel("Fecha"),
			widget.NewLabel("Descripción"),
			widget.NewLabel("Categoría"),
			widget.NewLabel("Tipo"),
			widget.NewLabel("Monto"),
			widget.NewLabel("Saldo"),
			widget.NewLabel("Adjunto"),
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

func (ui *UI) createTransactiontItem() fyne.CanvasObject {
	editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
	editBtn.Importance = widget.HighImportance

	voidBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
	voidBtn.Importance = widget.DangerImportance

	viewBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), nil)
	viewBtn.Importance = widget.LowImportance

	attachmentLink := componets.NewHoverableHyperlink("", nil, ui.mainWindow.Canvas())
	attachmentLink.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

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

	txtAmount := canvas.NewText("$1,200.00", color.Black)

	lblBalance := widget.NewLabel("$5,250.50")
	lblBalance.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)

	grid := container.NewGridWithColumns(9,
		lblTxNumber,
		lblDate,
		lblDescription,
		lblCategory,
		lblType,
		txtAmount,
		lblBalance,
		attachmentLink,
		container.NewHBox(
			editBtn,
			voidBtn,
			viewBtn,
		),
	)

	background := canvas.NewRectangle(color.Transparent)

	return container.NewStack(background, grid)
}

func (ui *UI) updateTransactionItem(i widget.ListItemID, o fyne.CanvasObject) {
	tx := ui.transactions.Data[i]

	stack := o.(*fyne.Container)
	background := stack.Objects[0].(*canvas.Rectangle)
	rowContainer := stack.Objects[1].(*fyne.Container)

	if tx.IsVoided {
		background.FillColor = color.NRGBA{R: 255, G: 0, B: 0, A: 40}
	} else {
		background.FillColor = color.Transparent
	}
	background.Refresh()

	rowContainer.Objects[0].(*widget.Label).SetText(tx.TransactionNumber)
	rowContainer.Objects[1].(*widget.Label).SetText(tx.TransactionDate.Format("01/02/2006"))
	rowContainer.Objects[2].(*widget.Label).SetText(helpers.PrepareForTruncation(tx.Description))

	if tx.Category != nil {
		rowContainer.Objects[3].(*widget.Label).SetText(tx.Category.Name)
		rowContainer.Objects[4].(*widget.Label).SetText(string(tx.Category.Type))
	} else {
		rowContainer.Objects[3].(*widget.Label).SetText("-")
		rowContainer.Objects[4].(*widget.Label).SetText("-")
	}

	amountText := rowContainer.Objects[5].(*canvas.Text)
	amountText.Text = fmt.Sprintf("%.2f", tx.Amount)
	if tx.Category != nil && tx.Category.Type == domain.Income {
		amountText.Text = "+ $" + amountText.Text
		amountText.Color = color.NRGBA{R: 0, G: 150, B: 0, A: 255} // Dark Green
	} else {
		amountText.Text = "- $" + amountText.Text
		amountText.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Dark Red
	}
	amountText.Refresh()

	rowContainer.Objects[6].(*widget.Label).SetText(fmt.Sprintf("$%.2f", tx.RunningBalance))

	attachmentLink := rowContainer.Objects[7].(*componets.HoverableHyperlink)
	if tx.AttachmentPath != nil && *tx.AttachmentPath != "" {
		fileName := filepath.Base(*tx.AttachmentPath)

		attachmentLink.SetText(helpers.PrepareForTruncation(fileName))
		attachmentLink.SetTooltip(fileName)

		dummyURL, _ := url.Parse("file://")
		attachmentLink.SetURL(dummyURL)

		attachmentLink.OnTapped = func() {
			previewDialog := transaction.NewPreviewDialog(ui.mainWindow, tx.AbsoluteAttachPath)
			previewDialog.Show()
		}
		attachmentLink.Show()
	} else {
		attachmentLink.SetText("-")
		attachmentLink.SetTooltip("")
		attachmentLink.OnTapped = nil
	}

	actionsContainer := rowContainer.Objects[8].(*fyne.Container)
	editBtn := actionsContainer.Objects[0].(*widget.Button)
	voidBtn := actionsContainer.Objects[1].(*widget.Button)
	viewBtn := actionsContainer.Objects[2].(*widget.Button)

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
			*ui.currentUser,
		)

		dialigHandler.Show()
	}

	voidBtn.OnTapped = func() {
		dialogHandler := transaction.NewVoidTransactionDialog(
			ui.mainWindow,
			ui.errorLogger,
			ui.Services.TxService,
			func() {
				ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
			},
			tx.ID,
			*ui.currentUser,
		)

		dialogHandler.Show()
	}

	viewBtn.OnTapped = func() {
		detailsDialog := transaction.NewDetailsDialog(ui.mainWindow, &tx)
		detailsDialog.Show()
	}

	if tx.IsVoided || tx.VoidsTransactionID != nil ||
		strings.Contains(tx.Category.Name, "Ajuste") {
		voidBtn.Hide()
		editBtn.Hide()
		viewBtn.Show()
	} else {
		voidBtn.Show()
		editBtn.Show()
		viewBtn.Hide()
	}
}

func (ui *UI) loadTransactions(page int, pageSize int) {
	if ui.selectedAccountID == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filters := ui.currentTransactionFilters

	result, err := ui.Services.TxService.FindTransactionsByAccount(
		ctx,
		ui.selectedAccountID,
		page,
		pageSize,
		filters,
		&ui.transactionSearchText)
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
	ui.transactionSearchText = filter
	ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
}

func (ui *UI) loadAccountsForTx() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

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

	ui.accountSelector.Options = accountNames
	if len(ui.accounts) > 0 {
		ui.selectedAccountID = ui.accounts[0].ID
		ui.accountSelector.SetSelected(ui.accounts[0].Name)
	}
	fyne.Do(func() {
		ui.accountSelector.Refresh()
	})
	go ui.loadTransactions(1, ui.transactionPaginator.GetPageSize())
}

func (ui *UI) generateReportFile(format string, outputPath string) {
	componets.HandleLongRunningOperation(ui.mainWindow, "Generando Reporte de Transacciones...", func(ctx context.Context) error {
		transactions, err := ui.Services.TxService.FindAllTransactionsByAccount(
			ctx,
			ui.selectedAccountID,
			ui.currentTransactionFilters,
		)
		if err != nil {
			return err
		}
		return ui.Services.ReportService.GenerateReportFile(ctx, format, transactions, outputPath, ui.currentUser)
	})
}

func (ui *UI) generateDailyReportFile(format string, outputPath string) {
	componets.HandleLongRunningOperation(ui.mainWindow, "Generando Reporte Diario...", func(ctx context.Context) error {
		report, err := ui.Services.ReportService.GenerateDailyReport(ctx, ui.selectedAccountID)
		if err != nil {
			return err
		}
		return ui.Services.ReportService.GenerateDailyReportFile(ctx, report, outputPath, format, ui.currentUser)
	})
}
