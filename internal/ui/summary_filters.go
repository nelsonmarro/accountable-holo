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
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
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
	ui.summaryStartDateEntry = componets.NewLatinDateEntry(ui.mainWindow)
	ui.summaryEndDateEntry = componets.NewLatinDateEntry(ui.mainWindow)

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	ui.summaryStartDateEntry.SetDate(firstOfMonth)
	ui.summaryEndDateEntry.SetDate(now)

	ui.summaryDateRangeSelect = widget.NewSelect(dateRangeOptions, nil)
	ui.summaryDateRangeSelect.SetSelected(dateRangeOptions[0])

	ui.summaryAccountSelect = widget.NewSelect([]string{}, nil)

	// Callback para visibilidad de fechas (usado también en makeFilterBar)
	ui.summaryDateRangeSelect.OnChanged = func(selected string) {
		// La lógica visual se maneja en el contenedor padre
	}

	generateBtn := widget.NewButtonWithIcon("Generar Resumen", theme.ViewRefreshIcon(), func() {
		go ui.generateSummary()
	})
	generateBtn.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabel("Filtros Inicializados"), // Placeholder invisible logicamente
	)
}

func (ui *UI) loadAccountsForSummary() {
	if ui.currentUser == nil || !ui.currentUser.CanViewReports() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	accounts, err := ui.Services.AccService.GetAllAccounts(ctx)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	ui.accounts = accounts
	options := []string{"Todas las Cuentas"}
	for _, acc := range accounts {
		options = append(options, acc.Name)
	}

	fyne.Do(func() {
		ui.summaryAccountSelect.Options = options
		ui.summaryAccountSelect.SetSelected(options[0])
		ui.summaryAccountSelect.Refresh()
	})

	go ui.generateSummary()
}

func (ui *UI) generateSummary() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	selectedRange := ui.summaryDateRangeSelect.Selected
	var startDate, endDate time.Time

	if selectedRange == "Personalizado" {
		if ui.summaryStartDateEntry.Date == nil || ui.summaryEndDateEntry.Date == nil {
			dialog.ShowError(fmt.Errorf("por favor seleccione ambas fechas"), ui.mainWindow)
			return
		}
		startDate = *ui.summaryStartDateEntry.Date
		endDate = *ui.summaryEndDateEntry.Date
	} else {
		startDate, endDate = getDatesForRange(selectedRange)
	}

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

	summary, err := ui.Services.ReportService.GetFinancialSummary(ctx, startDate, endDate, accountID)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	budgetStatuses, err := ui.Services.ReportService.GetBudgetOverview(ctx, startDate, endDate)
	if err != nil {
		ui.errorLogger.Printf("Error getting budget overview: %v", err)
	}

	fyne.Do(func() {
		updateMetricText(ui.summaryTotalIncome, summary.TotalIncome, domain.Income)
		updateMetricText(ui.summaryTotalExpenses, summary.TotalExpenses, domain.Outcome)
		updateMetricText(ui.summaryNetProfitLoss, summary.NetProfitLoss, "Net")

		// Update Charts using NEW Renderer
		ui.summaryChartsContainer.Objects = nil

		// Chart 1: Income vs Expense (Bar Chart Image)
		barChartImg := componets.RenderIncomeExpenseChart(summary.TotalIncome, summary.TotalExpenses)
		chart1Container := container.NewVBox(
			widget.NewLabelWithStyle("Comparativa Global", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			container.NewCenter(barChartImg),
		)

		// Chart 2: Expenses Breakdown (Pie Chart Image)
		pieChartImg := componets.RenderCategoryPieChart(summary.ExpensesByCategory)

		// Botón ver más (si hay datos)
		var chart2Objects []fyne.CanvasObject
		chart2Objects = append(chart2Objects, widget.NewLabelWithStyle("Distribución de Gastos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
		chart2Objects = append(chart2Objects, container.NewCenter(pieChartImg))

		if len(summary.ExpensesByCategory) > 0 {
			btn := widget.NewButton("Ver Detalle", func() {
				ui.showFullCategoryList(summary.ExpensesByCategory, summary.TotalExpenses)
			})
			btn.Importance = widget.LowImportance
			chart2Objects = append(chart2Objects, container.NewCenter(btn))
		}

		chart2Container := container.NewVBox(chart2Objects...)

		ui.summaryChartsContainer.Add(container.NewPadded(chart1Container))
		ui.summaryChartsContainer.Add(container.NewPadded(chart2Container))
		ui.summaryChartsContainer.Refresh()

		// Update Budget (Native List)
		ui.summaryBudgetContainer.Objects = nil
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
	// Revertimos a una lista simple nativa para el detalle "Ver Más",
	// ya que el gráfico de pastel es una imagen estática.
	content := container.NewVBox()
	for _, item := range data {
		percent := item.Amount.Div(total).Mul(decimal.NewFromFloat(100))
		percentFloat, _ := percent.Float64()
		lbl := fmt.Sprintf("%s: $%s (%.1f%%)", item.CategoryName, item.Amount.StringFixed(2), percentFloat)
		content.Add(widget.NewLabel(lbl))
	}

	scroll := container.NewVScroll(container.NewPadded(content))
	scroll.SetMinSize(fyne.NewSize(400, 500))

	dialog.NewCustom("Detalle de Gastos", "Cerrar", scroll, ui.mainWindow).Show()
}

func (ui *UI) showFullBudgetList(statuses []domain.BudgetStatus) {
	fullList := componets.NewBudgetStatusChart(statuses, nil)
	scroll := container.NewVScroll(container.NewPadded(fullList))
	scroll.SetMinSize(fyne.NewSize(600, 500))
	dialog.NewCustom("Detalle Completo de Presupuestos", "Cerrar", scroll, ui.mainWindow).Show()
}

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

