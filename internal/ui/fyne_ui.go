// Package ui provides the UI implementation for the application.
package ui

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// The UI struct holds the dependencies and state for the Fyne UI.
type UI struct {
	// ---- Dependencies ----
	accService AccountService
	catService CategoryService

	// ---- Fyne App Objects ----
	app        fyne.App
	mainWindow fyne.Window

	// ---- UI widgets (State) ----
	accountList *widget.List
	accounts    []domain.Account

	categoryList *widget.List
	categories   *domain.PaginatedResult[domain.Category]

	// ---- Debug ----
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewUI is the constructor for the UI struct.
func NewUI(
	accService AccountService,
	catService CategoryService,
) *UI {
	return &UI{
		accService:  accService,
		catService:  catService,
		infoLogger:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
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
		container.NewTabItemWithIcon("Cuentas", accountIcon, ui.makeAccountTab()),
		container.NewTabItemWithIcon("Finanzas", transactionIcon, ui.makeFinanceTab()),
		container.NewTabItemWithIcon("Reportes", reportIcon, widget.NewLabel("Reports")),
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
