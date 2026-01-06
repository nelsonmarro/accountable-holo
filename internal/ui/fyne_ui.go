// Package ui provides the Fyne-based user interface for the Accountable Holo application.
package ui

import (
	"context"
	"log"

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
	ReportService ReportService
	UserService   UserService
	RecurService  RecurringTransactionService
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
		mainWindow.Close()
		ui.openLoginWindow()
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

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Resumen Financiero", reportIcon, summaryTabContent),
		container.NewTabItemWithIcon("Cuentas", accountIcon, accountsTabContent),
		container.NewTabItemWithIcon("Transacciones", transactionIcon, txTabContent),
	)

	if ui.currentUser.Role == domain.AdminRole {
		userTabContent := widget.NewLabel("Cargando Usuarios...")
		tabs.Append(container.NewTabItemWithIcon("Usuarios", theme.AccountIcon(), userTabContent))
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
		// Use a detached context or one with sufficient timeout
		ctx := context.Background()
		// We use *ui.currentUser because we are already logged in
		if ui.currentUser != nil {
			err := ui.Services.RecurService.ProcessPendingRecurrences(ctx, *ui.currentUser)
			if err != nil {
				ui.errorLogger.Printf("Failed to process recurring transactions: %v", err)
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
					lbl.Text == "Cargando Usuarios..."
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
