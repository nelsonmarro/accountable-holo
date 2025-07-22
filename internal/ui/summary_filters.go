package ui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

var dateRangeOptions = []string{
	"Este Mes",
	"Mes Pasado",
	"Este Trimestre",
	"Este Año",
}

func (ui *UI) makeFilterCard() fyne.CanvasObject {
	ui.summaryDateRangeSelect = widget.NewSelect(dateRangeOptions, nil)
	ui.summaryDateRangeSelect.SetSelected(dateRangeOptions[0]) // Default to "This Month"

	ui.summaryAccountSelect = widget.NewSelect([]string{}, nil)
	go ui.loadAccountsForSummary()

	generateBtn := widget.NewButtonWithIcon("Generar Resumen", theme.ViewRefreshIcon(), func() {
		ui.generateSummary()
	})
	generateBtn.Importance = widget.HighImportance

	form := widget.NewForm(
		widget.NewFormItem("Rango de Fecha", ui.summaryDateRangeSelect),
		widget.NewFormItem("Cuenta", ui.summaryAccountSelect),
	)

	return container.NewBorder(
		nil,
		container.NewPadded(generateBtn),
		nil,
		nil,
		form,
	)
}

func (ui *UI) loadAccountsForSummary() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	accounts, err := ui.Services.AccService.GetAllAccounts(ctx)
	if err != nil {
		dialog.ShowError(err, ui.mainWindow)
		return
	}

	options := []string{"Todas las Cuentas"}
	for _, acc := range accounts {
		options = append(options, acc.Name)
	}

	fyne.Do(func() {
		ui.summaryAccountSelect.Options = options
		ui.summaryAccountSelect.SetSelected(options[0])
		ui.summaryAccountSelect.Refresh()
		// Trigger initial summary generation
		ui.generateSummary()
	})
}

func (ui *UI) generateSummary() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get selected date range
	selectedRange := ui.summaryDateRangeSelect.Selected
	startDate, endDate := getDatesForRange(selectedRange)

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
	go func() {
		summary, err := ui.Services.ReportService.GenerateFinancialSummary(ctx, startDate, endDate, accountID)
		if err != nil {
			dialog.ShowError(err, ui.mainWindow)
			return
		}

		// Update the UI
		fyne.Do(func() {
			updateMetricLabel(ui.summaryTotalIncome, summary.TotalIncome, domain.Income)
			updateMetricLabel(ui.summaryTotalExpenses, summary.TotalExpenses, domain.Outcome)
			updateMetricLabel(ui.summaryNetProfitLoss, summary.NetProfitLoss, "")
		})
	}()
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
