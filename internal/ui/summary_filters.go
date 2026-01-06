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
	"github.com/shopspring/decimal"
)

var dateRangeOptions = []string{
	"Este Mes",
	"Mes Pasado",
	"Este Trimestre",
	"Este Año",
	"Personalizado",
}

// makeFilterCard creates a card with filters for generating financial summaries.
func (ui *UI) makeFilterCard() fyne.CanvasObject {
	ui.summaryStartDateEntry = widget.NewDateEntry()
	ui.summaryEndDateEntry = widget.NewDateEntry()

	// Default to current month for custom entries
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	ui.summaryStartDateEntry.SetDate(&firstOfMonth)
	ui.summaryEndDateEntry.SetDate(&now)

	ui.summaryDateRangeSelect = widget.NewSelect(dateRangeOptions, nil)
	ui.summaryDateRangeSelect.SetSelected(dateRangeOptions[0]) // Default to "This Month"

	ui.summaryAccountSelect = widget.NewSelect([]string{}, nil)

	generateBtn := widget.NewButtonWithIcon("Generar Resumen", theme.ViewRefreshIcon(), func() {
		go ui.generateSummary()
	})
	generateBtn.Importance = widget.HighImportance

	// We wrap the custom dates in a container to manage their visibility together
	// though they are also FormItems.
	// Fyne's Form doesn't hide labels when widgets are hidden, so we'll use a VBox
	// for the optional part.

	form := widget.NewForm(
		widget.NewFormItem("Rango de Fecha", ui.summaryDateRangeSelect),
	)

	customDatesForm := widget.NewForm(
		widget.NewFormItem("Fecha Inicial", ui.summaryStartDateEntry),
		widget.NewFormItem("Fecha Final", ui.summaryEndDateEntry),
	)
	customDatesForm.Hide()

	// Update the select callback to handle the form visibility
	ui.summaryDateRangeSelect.OnChanged = func(selected string) {
		if selected == "Personalizado" {
			customDatesForm.Show()
		} else {
			customDatesForm.Hide()
		}
	}

	accountForm := widget.NewForm(
		widget.NewFormItem("Cuenta", ui.summaryAccountSelect),
	)

	return container.NewBorder(
		nil,
		container.NewPadded(generateBtn),
		nil,
		nil,
		container.NewVBox(
			form,
			customDatesForm,
			accountForm,
		),
	)
}

// loadAccountsForSummary fetches all accounts and populates the account selection dropdown.
func (ui *UI) loadAccountsForSummary() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	accounts, err := ui.Services.AccService.GetAllAccounts(ctx)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.accounts = accounts // Keep a reference to the accounts slice in UI struct
	options := []string{"Todas las Cuentas"}
	for _, acc := range accounts {
		options = append(options, acc.Name)
	}

	fyne.Do(func() {
		ui.summaryAccountSelect.Options = options
		ui.summaryAccountSelect.SetSelected(options[0])
		ui.summaryAccountSelect.Refresh()
	})

	// Trigger initial summary generation
	go ui.generateSummary()
}

func (ui *UI) generateSummary() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get selected date range
	selectedRange := ui.summaryDateRangeSelect.Selected
	var startDate, endDate time.Time

	if selectedRange == "Personalizado" {
		if ui.summaryStartDateEntry.Date == nil || ui.summaryEndDateEntry.Date == nil {
			dialog.ShowError(fmt.Errorf("por favor seleccione ambas fechas para el rango personalizado"), ui.mainWindow)
			return
		}
		startDate = *ui.summaryStartDateEntry.Date
		endDate = *ui.summaryEndDateEntry.Date
	} else {
		startDate, endDate = getDatesForRange(selectedRange)
	}

	// Get selected account
	selectedAccountName := ui.summaryAccountSelect.Selected
	var accountID *int
	if selectedAccountName != "Todas las Cuentas" {
		for _, acc := range ui.accounts {
			if acc.Name == selectedAccountName {
				id := acc.ID
				accountID = &id
				break
			}
		}
	}

	// Call the service
	summary, err := ui.Services.ReportService.GetFinancialSummary(ctx, startDate, endDate, accountID)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	budgetStatuses, err := ui.Services.ReportService.GetBudgetOverview(ctx, startDate, endDate)
	if err != nil {
		ui.errorLogger.Printf("Error getting budget overview: %v", err)
		// We don't block the UI, just show an error or empty budget section
	}

	// Update the UI
	fyne.Do(func() {
		updateMetricText(ui.summaryTotalIncome, summary.TotalIncome, domain.Income)
		updateMetricText(ui.summaryTotalExpenses, summary.TotalExpenses, domain.Outcome)
		updateMetricText(ui.summaryNetProfitLoss, summary.NetProfitLoss, "Net")

		// Update Charts
		ui.summaryChartsContainer.Objects = nil // Clear previous charts

		// Chart 1: Income vs Expense with Title
		barChart := componets.NewIncomeExpenseChart(summary.TotalIncome, summary.TotalExpenses)
		chart1Container := container.NewVBox(
			widget.NewLabelWithStyle("Comparativa Global: Ingresos vs Egresos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			barChart,
		)

		// Chart 2: Expenses Breakdown with Title
		expenseChart := componets.NewCategoryBreakdownChart(
			summary.ExpensesByCategory,
			summary.TotalExpenses,
			func() {
				ui.showFullCategoryList(summary.ExpensesByCategory, summary.TotalExpenses)
			},
		)
		chart2Container := container.NewVBox(
			widget.NewLabelWithStyle("Distribución de Gastos: ¿A dónde va el dinero?", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			expenseChart,
		)

		ui.summaryChartsContainer.Add(container.NewPadded(chart1Container))
		ui.summaryChartsContainer.Add(container.NewPadded(chart2Container))
		ui.summaryChartsContainer.Refresh()

		// Update Budget
		ui.summaryBudgetContainer.Objects = nil
		ui.summaryBudgetContainer.Add(widget.NewLabelWithStyle("Seguimiento de Presupuestos: Controla tus límites de gasto mensual.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))

		budgetChart := componets.NewBudgetStatusChart(
			budgetStatuses,
			func() {
				ui.showFullBudgetList(budgetStatuses)
			},
		)
		ui.summaryBudgetContainer.Add(budgetChart)
		ui.summaryBudgetContainer.Refresh()
	})
}

func (ui *UI) showFullCategoryList(data []domain.CategoryAmount, total decimal.Decimal) {
	// Re-use the chart logic but without limit (pass nil callback) to generate the full list view
	// OR create a specific list view. Using the chart component is easier for consistency.
	fullList := componets.NewCategoryBreakdownChart(data, total, nil)

	scroll := container.NewVScroll(container.NewPadded(fullList))
	scroll.SetMinSize(fyne.NewSize(500, 400))

	dialog.NewCustom("Distribución Completa de Gastos", "Cerrar", scroll, ui.mainWindow).Show()
}

func (ui *UI) showFullBudgetList(statuses []domain.BudgetStatus) {
	// Re-use the chart component to show the full list
	fullList := componets.NewBudgetStatusChart(statuses, nil)

	scroll := container.NewVScroll(container.NewPadded(fullList))
	scroll.SetMinSize(fyne.NewSize(600, 500))

	dialog.NewCustom("Detalle Completo de Presupuestos", "Cerrar", scroll, ui.mainWindow).Show()
}

// getDatesForRange translates the selected string into start and end dates.
func getDatesForRange(r string) (start time.Time, end time.Time) {
	now := time.Now()
	year, month, _ := now.Date()

	switch r {
	case "Este Mes":
		start = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 1, -1)
	case "Mes Pasado":
		start = time.Date(year, month-1, 1, 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 1, -1)
	case "Este Trimestre":
		quarter := (int(month) - 1) / 3
		startMonth := time.Month(quarter*3 + 1)
		start = time.Date(year, startMonth, 1, 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 3, -1)
	case "Este Año":
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		end = time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)
	default:
		start = now
		end = now
	}
	return
}
