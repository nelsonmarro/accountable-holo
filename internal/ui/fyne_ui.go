// Package ui provides the UI implementation for the application.
package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

type Services struct {
	AccService    AccountService
	CatService    CategoryService
	TxService     TransactionService
	ReportService ReportService
}

// The UI struct holds the dependencies and state for the Fyne UI.
type UI struct {
	// ---- Dependencies ----
	Services *Services

	// ---- Fyne App Objects ----
	app        fyne.App
	mainWindow fyne.Window

	// ---- UI widgets (State) ----
	accountList *widget.List
	accounts    []domain.Account

	categoryList      *widget.List
	categoryPaginator *componets.Pagination
	categories        *domain.PaginatedResult[domain.Category]
	categoryFilter    string

	transactionList           *widget.List
	transactionPaginator      *componets.Pagination
	transactions              *domain.PaginatedResult[domain.Transaction]
	transactionFilter         string
	currentTransactionFilters domain.TransactionFilters
	accountSelector           *widget.Select
	selectedAccountID         int

	// ---- Summary Tab State ----
	summaryDateRangeSelect *widget.Select
	summaryAccountSelect   *widget.Select
	summaryTotalIncome     *canvas.Text
	summaryTotalExpenses   *canvas.Text
	summaryNetProfitLoss   *canvas.Text

	// ---- Debug ----
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func (ui *UI) generateReport(format string) {
	if ui.selectedAccountID == 0 {
		dialog.ShowInformation("Generar Reporte", "Por favor, seleccione una cuenta para generar el reporte.", ui.mainWindow)
		return
	}

	progress := dialog.NewProgressInfinite("Generando Reporte", "Por favor espere...", ui.mainWindow)
	progress.Show()

	go func() {
		defer progress.Hide()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout for report generation
		defer cancel()

		// Get the currently filtered transactions
		transactions, err := ui.Services.TxService.FindAllTransactionsByAccount(ctx, ui.selectedAccountID, ui.currentTransactionFilters)
		if err != nil {
			ui.errorLogger.Printf("Error fetching transactions for report: %v", err)
			dialog.ShowError(fmt.Errorf("Error al obtener transacciones para el reporte: %v", err), ui.mainWindow)
			return
		}

		// Determine output path (e.g., user's documents directory or a temp directory)
		// For now, let's use a temporary file.
		outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("report_%s.%s", time.Now().Format("20060102150405"), strings.ToLower(format)))

		err = ui.Services.ReportService.GenerateReportFile(ctx, format, transactions, outputPath)
		if err != nil {
			ui.errorLogger.Printf("Error generating report file: %v", err)
			dialog.ShowError(fmt.Errorf("Error al generar el archivo de reporte: %v", err), ui.mainWindow)
			return
		}

		ui.infoLogger.Printf("Report generated successfully at: %s", outputPath)
		dialog.ShowInformation("Reporte Generado", fmt.Sprintf("Reporte en formato %s generado exitosamente en:\n%s", format, outputPath), ui.mainWindow)
	}()
}

// NewUI is the constructor for the UI struct.
func NewUI(services *Services) *UI {
	return &UI{
		Services:    services,
		infoLogger:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		categories:  &domain.PaginatedResult[domain.Category]{},
	}
}

// Init creates the Fyne app and window objects. This is where the Fyne-specific
func (ui *UI) Init(a fyne.App) {
	ui.app = a
	ui.app.Settings().SetTheme(NewAppTheme())
	ui.mainWindow = ui.app.NewWindow("Accountable Holo")
}

// buildMainUI creates all the main UI components and sets them on the window.
func (ui *UI) buildMainUI() {
	accountIcon := NewThemeAwareResource(resourceAccountstabiconlightPng, resourceAccountstabicondarkPng)
	transactionIcon := NewThemeAwareResource(resourceTransactionstabiconlightPng, resourceTransactiontabicondarkPng)
	reportIcon := NewThemeAwareResource(resourceReportstabiconlightPng, resourceReportstabicondarkPng)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Resumen Financiero", reportIcon, ui.makeSummaryTab()),
		container.NewTabItemWithIcon("Cuentas", accountIcon, ui.makeAccountTab()),
		container.NewTabItemWithIcon("Transacciones", transactionIcon, ui.makeFinancesTab()),
		container.NewTabItemWithIcon("Reportes", reportIcon, widget.NewLabel("Reportes")),
	)

	ui.mainWindow.SetContent(tabs)
	ui.mainWindow.Resize(fyne.NewSize(1280, 720))
	ui.mainWindow.CenterOnScreen()
	ui.mainWindow.SetMaster()
}

// Run now simply builds and then runs the application.
func (ui *UI) Run() {
	ui.buildMainUI()
	ui.mainWindow.ShowAndRun()
}
