// Package ui provides the Fyne-based user interface for the Accountable Holo application.
package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/licensing"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

type Services struct {
	AccService    AccountService
	CatService    CategoryService
	TxService     TransactionService
	UserService   UserService
	ReportService ReportService
	RecurService  RecurringTransactionService
	IssuerService IssuerService
	SriService    SriService
	TaxService    TaxPayerService // Added
}

// The UI struct holds the dependencies and state for the Fyne UI.
type UI struct {
	// ---- Dependencies ----
	Services *Services

	// ---- Fyne App Objects ----
	app        fyne.App
	mainWindow fyne.Window // Reference to the currently active window

	// ---- Auth State ----
	currentUser *domain.User

	// ---- UI widgets (State) ----
	userList *widget.List
	users    []domain.User

	accountList *widget.List
	accounts    []domain.Account

	categoryList      *widget.List
	categoryPaginator *componets.Pagination
	categories        *domain.PaginatedResult[domain.Category]
	categoryFilter    string

	// ---- TaxPayer State ----
	taxPayerList       *widget.List
	taxPayerPaginator  *componets.Pagination
	paginatedTaxPayers *domain.PaginatedResult[domain.TaxPayer]
	taxPayerSearchText string
	taxPayers          []domain.TaxPayer // Deprecated, but keeping for compatibility if needed

	transactionList           *widget.List
	transactionPaginator      *componets.Pagination
	transactions              *domain.PaginatedResult[domain.Transaction]
	transactionSearchText     string
	currentTransactionFilters domain.TransactionFilters
	accountSelector           *widget.Select
	selectedAccountID         int

	// ---- Summary Tab State ----
	summaryDateRangeSelect *widget.Select
	summaryStartDateEntry  *widget.DateEntry
	summaryEndDateEntry    *widget.DateEntry
	summaryAccountSelect   *widget.Select
	summaryTotalIncome     *canvas.Text
	summaryTotalExpenses   *canvas.Text
	summaryNetProfitLoss   *canvas.Text
	summaryChartsContainer *fyne.Container
	summaryBudgetContainer *fyne.Container

	// ---- Debug ----
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewUI(services *Services, infoLogger, errorLogger *log.Logger) *UI {
	return &UI{
		Services:    services,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		categories:  &domain.PaginatedResult[domain.Category]{},
	}
}

// Init sets the Fyne app object.
func (ui *UI) Init(a fyne.App) {
	ui.app = a
	ui.app.Settings().SetTheme(NewAppTheme())
}

// Run starts the application by opening the login window.
func (ui *UI) Run(licMgr *licensing.LicenseManager) {
	// Definimos la accion de exito (entrar al login)
	// onLicenseValid := func() {
	// 	ui.openLoginWindow()
	// }
	//
	// // Llamamos a la ventana de licencia
	// ui.ShowLincenseWindow(licMgr, onLicenseValid)
	ui.openLoginWindow()
	ui.app.Run()
}

func (ui *UI) openLoginWindow() {
	loginWindow := ui.app.NewWindow("Login - Accountable Holo")
	ui.mainWindow = loginWindow // Update the reference so dialogs work
	loginWindow.SetContent(ui.makeLoginUI(loginWindow))
	loginWindow.Resize(fyne.NewSize(400, 300))
	loginWindow.CenterOnScreen()
	loginWindow.Show()
}

func (ui *UI) openMainWindow() {
	mainWindow := ui.app.NewWindow("Accountable Holo")
	ui.mainWindow = mainWindow // Update the reference for dialogs

	// Create the menu with logout logic
	logoutItem := fyne.NewMenuItem("Cerrar Sesión", func() {
		ui.currentUser = nil
		ui.openLoginWindow()
		
		mainWindow.Hide()
		go func() {
			time.Sleep(100 * time.Millisecond)
			mainWindow.Close()
		}()
	})
	fileMenu := fyne.NewMenu("Sesión", logoutItem)
	mainWindow.SetMainMenu(fyne.NewMainMenu(fileMenu))

	// Build tabs
	accountIcon := NewThemeAwareResource(resourceAccountstabiconlightPng, resourceAccountstabicondarkPng)
	transactionIcon := NewThemeAwareResource(resourceTransactionstabiconlightPng, resourceTransactiontabicondarkPng)
	reportIcon := NewThemeAwareResource(resourceReportstabiconlightPng, resourceReportstabicondarkPng)

	// Summary Tab (Load immediately)
	summaryTabContent := ui.makeSummaryTab()

	// Placeholders
	accountsTabContent := widget.NewLabel("Cargando Cuentas...")
	txTabContent := widget.NewLabel("Cargando Transacciones...")
	taxPayerTabContent := widget.NewLabel("Cargando Clientes...")

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Resumen Financiero", reportIcon, summaryTabContent),
		container.NewTabItemWithIcon("Cuentas", accountIcon, accountsTabContent),
		container.NewTabItemWithIcon("Clientes", theme.AccountIcon(), taxPayerTabContent),
		container.NewTabItemWithIcon("Transacciones", transactionIcon, txTabContent),
	)

	if ui.currentUser.Role == domain.AdminRole {
		userTabContent := widget.NewLabel("Cargando Usuarios...")
		tabs.Append(container.NewTabItemWithIcon("Usuarios", theme.AccountIcon(), userTabContent))

		// Configuración SRI (Solo Admin)
		sriConfigContent := ui.makeSriConfigTab()
		tabs.Append(container.NewTabItemWithIcon("Configuración SRI", theme.SettingsIcon(), sriConfigContent))
	}

	ui.lazyLoadTabsContent(tabs)

	mainWindow.SetContent(tabs)
	mainWindow.Resize(fyne.NewSize(1280, 720))
	mainWindow.CenterOnScreen()
	mainWindow.Show()

	// Initial data load
	go ui.loadAccountsForSummary()

	// ---- Background Tasks (Recurrence) ----
	go func() {
		ctx := context.Background()
		if ui.currentUser != nil {
			err := ui.Services.RecurService.ProcessPendingRecurrences(ctx, *ui.currentUser)
			if err != nil {
				ui.errorLogger.Printf("Failed to process recurring transactions: %v", err)
			}
		}
	}()

	// ---- Background Worker (SRI Sync) ----
	go func() {
		ticker := time.NewTicker(2 * time.Minute) // Check every 2 minutes
		for range ticker.C {
			if ui.currentUser == nil {
				continue // Don't run if logged out
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			count, err := ui.Services.SriService.ProcessBackgroundSync(ctx)
			cancel()

			if err != nil {
				ui.errorLogger.Printf("SRI Sync Error: %v", err)
			}

			if count > 0 {
				ui.app.SendNotification(fyne.NewNotification("Facturación Electrónica", fmt.Sprintf("%d comprobantes han sido autorizados por el SRI.", count)))
				// Refresh list if active
				// Note: Ideally use a data binding or event bus, but direct refresh works for MVP
				if ui.transactionList != nil {
					ui.loadTransactions(ui.transactionPaginator.GetCurrentPage(), ui.transactionPaginator.GetPageSize())
				}
			}
		}
	}()
}

func (ui *UI) lazyLoadTabsContent(tabs *container.AppTabs) {
	tabs.OnSelected = func(item *container.TabItem) {
		// Helper to check if content is a placeholder label
		isPlaceholder := func(obj fyne.CanvasObject) bool {
			if lbl, ok := obj.(*widget.Label); ok {
				return lbl.Text == "Cargando Cuentas..." ||
					lbl.Text == "Cargando Transacciones..." ||
					lbl.Text == "Cargando Usuarios..." ||
					lbl.Text == "Cargando Clientes..."
			}
			return false
		}

		switch item.Text {
		case "Cuentas":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeAccountTab()
				tabs.Refresh()
			}
			if len(ui.accounts) == 0 {
				go ui.loadAccounts()
			}
		case "Clientes":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeTaxPayerTab()
				tabs.Refresh()
			}
			// Initial load is handled inside makeTaxPayerTab via goroutine
		case "Transacciones":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeFinancesTab()
				tabs.Refresh()
			}
			go ui.loadAccountsForTx()

			if ui.categories == nil || len(ui.categories.Data) == 0 {
				go ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
			}
		case "Usuarios":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeUserTab()
				tabs.Refresh()
			}
			if len(ui.users) == 0 {
				go ui.loadUsers()
			}
		}
	}
}
