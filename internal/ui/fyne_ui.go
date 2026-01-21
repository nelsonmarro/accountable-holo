// Package ui provides the Fyne-based user interface for the Verith application.
package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/licensing"
	"github.com/nelsonmarro/verith/internal/ui/componets"
	componets_user "github.com/nelsonmarro/verith/internal/ui/componets/user"
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
	// 1. Verificar Estado de Licencia
	statusData, err := licMgr.CheckStatus()
	if err != nil {
		ui.errorLogger.Printf("Error checking license: %v", err)
		// Fallback safe: Show license window if check fails critically
		ui.ShowLincenseWindow(licMgr, func() { ui.openLoginWindow() })
		ui.app.Run()
		return
	}

	// 2. Lógica de Bloqueo
	if statusData.Status == licensing.StatusExpired {
		ui.ShowLincenseWindow(licMgr, func() {
			ui.openLoginWindow()
		})
	} else {
		// Active o Trial -> Permitir acceso
		// Si es trial, podríamos mostrar un toast/notificación, pero por ahora directo al login
		ui.openLoginWindow()
	}

	ui.app.Run()
}

func (ui *UI) openLoginWindow() {
	// Verificar si es la primera ejecución (sin usuarios)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hasUsers, err := ui.Services.UserService.HasUsers(ctx)
	if err != nil {
		ui.errorLogger.Printf("Error checking for users: %v", err)
		// Si falla, asumimos que hay usuarios para no bloquear, o mostramos error fatal.
		// Mejor mostrar error en UI temporalmente
	}

	if !hasUsers {
		// Flujo de Registro Inicial
		registerWindow := ui.app.NewWindow("Bienvenido - Crear Administrador")
		ui.mainWindow = registerWindow

		onSuccess := func() {
			registerWindow.Close()
			// Al terminar el registro, volvemos a llamar a esta función.
			// Como ya habrá usuarios, irá al else (Login).
			ui.openLoginWindow()
		}

		regDialog := componets_user.NewRegistrationDialog(registerWindow, ui.Services.UserService, ui.infoLogger, onSuccess)
		regDialog.Show()

		registerWindow.Resize(fyne.NewSize(500, 400))
		registerWindow.CenterOnScreen()
		registerWindow.Show()
		return
	}

	// Flujo Normal de Login
	loginWindow := ui.app.NewWindow("Login - Verith")
	ui.mainWindow = loginWindow // Update the reference so dialogs work
	loginWindow.SetContent(ui.makeLoginUI(loginWindow))
	loginWindow.Resize(fyne.NewSize(400, 300))
	loginWindow.CenterOnScreen()
	loginWindow.Show()
}

func (ui *UI) openMainWindow() {
	mainWindow := ui.app.NewWindow("Verith")
	ui.mainWindow = mainWindow // Update the reference for dialogs

	// --- License Check for UI ---
	// Need a license manager instance here to check status for UI elements
	configDir, _ := os.UserConfigDir()
	licensePath := filepath.Join(configDir, "Verith") // Use same path logic as main
	licMgr := licensing.NewLicenseManager(licensePath)
	licenseData, _ := licMgr.CheckStatus()

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

	// Add License Management Item
	licenseItem := fyne.NewMenuItem("Gestionar Licencia", func() {
		ui.ShowLincenseWindow(licMgr, func() {}) // Callback empty as we are already inside
	})

	fileMenu := fyne.NewMenu("Sesión", licenseItem, fyne.NewMenuItemSeparator(), logoutItem)
	mainWindow.SetMainMenu(fyne.NewMainMenu(fileMenu))

	// Build tabs
	accountIcon := NewThemeAwareResource(resourceAccountstabiconlightPng, resourceAccountstabicondarkPng)
	transactionIcon := NewThemeAwareResource(resourceTransactionstabiconlightPng, resourceTransactiontabicondarkPng)
	reportIcon := NewThemeAwareResource(resourceReportstabiconlightPng, resourceReportstabicondarkPng)

	tabs := container.NewAppTabs()

	// 1. Resumen Financiero (Admin y Supervisor)
	if ui.currentUser.CanViewReports() {
		summaryTabContent := ui.makeSummaryTab()
		tabs.Append(container.NewTabItemWithIcon("Resumen Financiero", reportIcon, summaryTabContent))
	}

	// 2. Cuentas (Todos)
	accountsTabContent := widget.NewLabel("Cargando Cuentas...")
	tabs.Append(container.NewTabItemWithIcon("Cuentas", accountIcon, accountsTabContent))

	// 3. Clientes (Todos)
	taxPayerTabContent := widget.NewLabel("Cargando Clientes...")
	tabs.Append(container.NewTabItemWithIcon("Clientes", theme.AccountIcon(), taxPayerTabContent))

	// 4. Transacciones (Todos)
	txTabContent := widget.NewLabel("Cargando Transacciones...")
	tabs.Append(container.NewTabItemWithIcon("Transacciones", transactionIcon, txTabContent))

	// 5. Usuarios (Solo Admin)
	if ui.currentUser.CanManageUsers() {
		userTabContent := widget.NewLabel("Cargando Usuarios...")
		tabs.Append(container.NewTabItemWithIcon("Usuarios", theme.AccountIcon(), userTabContent))
	}

	// 6. Configuración SRI (Solo Admin)
	if ui.currentUser.CanConfigureSystem() {
		sriConfigContent := ui.makeSriConfigTab()
		tabs.Append(container.NewTabItemWithIcon("Configuración SRI", theme.SettingsIcon(), sriConfigContent))
	}

	ui.lazyLoadTabsContent(tabs)

	// --- Trial Banner ---
	var mainContent fyne.CanvasObject
	if licenseData.Status == licensing.StatusTrial {
		daysLeft := max(15-int(time.Since(licenseData.InstallDate).Hours()/24), 0)

		bannerLabel := widget.NewLabel(fmt.Sprintf("⚠ Modo de Prueba: %d días restantes.", daysLeft))
		bannerLabel.TextStyle = fyne.TextStyle{Bold: true}

		activateBtn := widget.NewButton("Activar Ahora", func() {
			ui.ShowLincenseWindow(licMgr, func() {})
		})
		activateBtn.Importance = widget.HighImportance

		// Create a colored background container if possible, or just a VBox
		banner := container.NewHBox(
			widget.NewIcon(theme.WarningIcon()),
			bannerLabel,
			layout.NewSpacer(),
			activateBtn,
		)

		mainContent = container.NewBorder(container.NewPadded(banner), nil, nil, nil, tabs)
	} else {
		mainContent = tabs
	}

	mainWindow.SetContent(mainContent)
	mainWindow.Resize(fyne.NewSize(1280, 720))
	mainWindow.CenterOnScreen()
	mainWindow.Show()
	// Chequeo de Configuración Inicial (Onboarding)
	go func() {
		// Damos un momento para que la ventana se renderice
		time.Sleep(500 * time.Millisecond)

		ctx := context.Background()
		issuer, err := ui.Services.IssuerService.GetIssuerConfig(ctx)

		if err == nil && issuer == nil {
			// No hay configuración
			fyne.Do(func() {
				if ui.currentUser.Role == domain.RoleAdmin {
					// Admin: Mostrar Wizard
					content := ui.makeSriConfigTab()
					d := dialog.NewCustom("Bienvenido a Verith - Configuración Inicial", "Cerrar", content, mainWindow)
					d.Resize(fyne.NewSize(850, 700))
					d.Show()

					dialog.ShowInformation("Primeros Pasos",
						"Para comenzar a facturar, por favor configura los datos de tu empresa y firma electrónica.", mainWindow)
				} else {
					// Standard: Mostrar Alerta
					dialog.ShowInformation("Configuración Pendiente",
						"El sistema aún no está configurado para facturación electrónica.\nPor favor, solicita a un administrador que configure los datos de la empresa.", mainWindow)
				}
			})
		}
	}()

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
