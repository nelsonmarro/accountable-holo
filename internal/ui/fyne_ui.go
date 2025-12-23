// Package ui provides the Fyne-based user interface for the Accountable Holo application.
package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

type Services struct {
	AccService    AccountService
	CatService    CategoryService
	TxService     TransactionService
	ReportService ReportService
	UserService   UserService
}

// The UI struct holds the dependencies and state for the Fyne UI.
type UI struct {
	// ---- Dependencies ----
	Services *Services

	// ---- Fyne App Objects ----
	app        fyne.App
	mainWindow fyne.Window

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
	summaryAccountSelect   *widget.Select
	summaryTotalIncome     *canvas.Text
	summaryTotalExpenses   *canvas.Text
	summaryNetProfitLoss   *canvas.Text

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

// Init creates the Fyne app and window objects. This is where the Fyne-specific
func (ui *UI) Init(a fyne.App) {
	ui.app = a
	ui.app.Settings().SetTheme(NewAppTheme())
	ui.mainWindow = ui.app.NewWindow("Accountable Holo")
}

// buildMainUI creates the main application layout and menu.
func (ui *UI) buildMainUI() *container.AppTabs {
	tabs := container.NewAppTabs()
	ui.mainWindow.SetMainMenu(ui.makeMainMenu())
	return tabs
}

func lazyLoadDbCalls(tabs *container.AppTabs, ui *UI) {
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
				tabs.Refresh() // Force redraw of the tab content
			}
			if ui.accounts == nil || len(ui.accounts) == 0 {
				go ui.loadAccounts()
			}
		case "Transacciones":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeFinancesTab()
				tabs.Refresh()
			}
			if ui.transactions == nil || len(ui.transactions.Data) == 0 {
				go ui.loadAccountsForTx()
			}

			if ui.categories == nil || len(ui.categories.Data) == 0 {
				go ui.loadCategories(1, ui.categoryPaginator.GetPageSize())
			}
		case "Usuarios":
			if isPlaceholder(item.Content) {
				item.Content = ui.makeUserTab()
				tabs.Refresh()
			}
			if ui.users == nil || len(ui.users) == 0 {
				go ui.loadUsers()
			}
		}
	}
}

// Run now simply builds and then runs the application.
func (ui *UI) Run() {
	ui.mainWindow.SetContent(ui.makeLoginUI())
	ui.mainWindow.Resize(fyne.NewSize(500, 300))
	ui.mainWindow.CenterOnScreen()
	ui.mainWindow.ShowAndRun()
}

func (ui *UI) makeMainMenu() *fyne.MainMenu {
	logoutItem := fyne.NewMenuItem("Cerrar Sesión", func() {
		ui.currentUser = nil
		ui.mainWindow.SetMainMenu(nil)
		ui.mainWindow.SetContent(ui.makeLoginUI())
		ui.mainWindow.Resize(fyne.NewSize(500, 300))
		ui.mainWindow.CenterOnScreen()
	})

	fileMenu := fyne.NewMenu("Sesión", logoutItem)

	return fyne.NewMainMenu(fileMenu)
}
