// Package ui provides the UI implementation for the application.
package ui

import (
	"context"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	)

	if ui.currentUser.Role == domain.AdminRole {
		tabs.Append(container.NewTabItemWithIcon("Users", theme.AccountIcon(), ui.makeUserTab()))
	}

	ui.mainWindow.SetContent(tabs)
	ui.mainWindow.Resize(fyne.NewSize(1280, 720))
	ui.mainWindow.CenterOnScreen()
	ui.mainWindow.SetMaster()
}

// Run now simply builds and then runs the application.
func (ui *UI) Run() {
	ui.mainWindow.SetContent(ui.makeLoginUI())
	ui.mainWindow.Resize(fyne.NewSize(400, 200))
	ui.mainWindow.CenterOnScreen()
	ui.mainWindow.ShowAndRun()
}

func (ui *UI) makeLoginUI() fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			user, err := ui.Services.UserService.Login(ctx, usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				dialog.ShowError(err, ui.mainWindow)
				return
			}

			ui.currentUser = user
			ui.buildMainUI()
		},
	}

	return container.NewCenter(loginForm)
}
