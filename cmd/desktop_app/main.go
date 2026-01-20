package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/nelsonmarro/verith/config"
	"github.com/nelsonmarro/verith/internal/application/report"
	"github.com/nelsonmarro/verith/internal/application/service"
	"github.com/nelsonmarro/verith/internal/infrastructure/database"
	persistence "github.com/nelsonmarro/verith/internal/infrastructure/persistence"
	"github.com/nelsonmarro/verith/internal/infrastructure/storage"
	"github.com/nelsonmarro/verith/internal/licensing"
	"github.com/nelsonmarro/verith/internal/logging"
	"github.com/nelsonmarro/verith/internal/security"
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/nelsonmarro/verith/internal/ui"
)

// Esta variable la llenará el compilador (linker)
var ResendAPIKeyEncrypted string

func main() {
	// ---- Logging ----
	infoLogger, errorLogger, err := logging.Init()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// --- Panic Recovery ---
	// This ensures that if the app crashes, we capture the error in the log file.
	defer func() {
		if r := recover(); r != nil {
			errorLogger.Printf("CRITICAL ERROR (PANIC): %v\nStack Trace:\n%s", r, debug.Stack())
		}
	}()

	infoLogger.Println("Application starting...")

	conf, err := config.LoadConfig("config")
	if err != nil {
		errorLogger.Fatalf("failed to load configuration: %v", err)
	}
	infoLogger.Println("Config loaded.")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ---- Infrastructure (Database) ----
	infoLogger.Println("Connecting to database...")
	pool, err := database.Connect(ctx, conf)
	if err != nil {
		errorLogger.Fatalf("failed to connect to the database: %v", err)
	}
	defer pool.Close()
	infoLogger.Println("Connected to the database successfully")

	// ---- UI (Fyne) ----
	infoLogger.Println("Initializing Fyne App...")
	// 1. Create the Fyne App first.
	// a := app.NewWithID("51af2ee4-c61c-4608-a3f1-d8576343af14") removed
	infoLogger.Println("Fyne App initialized.")

	// ---- Infrastructure (Storage) ----
	infoLogger.Println("Initializing Storage Service...")
	storageService, err := storage.NewLocalStorageService(conf.Storage.AttachmentPath)
	if err != nil {
		errorLogger.Fatalf("failed to create storage service: %v", err)
	}
	infoLogger.Println("Storage Service initialized.")

	// ---- Infrastructure (Repositories) ----
	accRepo := persistence.NewAccountRepository(pool)
	catRepo := persistence.NewCategoryRepository(pool)
	txRepo := persistence.NewTransactionRepository(pool)
	reportRepo := persistence.NewReportRepository(pool)
	userRepo := persistence.NewUserRepository(pool)
	recurRepo := persistence.NewRecurringTransactionRepository(pool)
	issuerRepo := persistence.NewIssuerRepository(pool)
	receiptRepo := persistence.NewElectronicReceiptRepository(pool)
	clientRepo := persistence.NewTaxPayerRepository(pool)
	emissionRepo := persistence.NewEmissionPointRepository(pool)

	// ---- Application (Report Generators) ----
	csvGen := report.NewCSVReportGenerator()
	pdfGen := report.NewPDFReportGenerator()

	// ---- SRI (Client & Service) ----
	sriClient := sri.NewSoapClient()

	// ---- Application (Services) ----
	accService := service.NewAccountService(accRepo)
	catService := service.NewCategoryService(catRepo)
	txService := service.NewTransactionService(txRepo, storageService, accService)
	userService := service.NewUserService(userRepo)
	reportService := service.NewReportService(reportRepo, txRepo, catRepo, csvGen, pdfGen)
	recurService := service.NewRecurringTransactionService(recurRepo, txRepo, infoLogger)
	issuerService := service.NewIssuerService(issuerRepo, emissionRepo)
	taxService := service.NewTaxPayerService(clientRepo)

	// Decodificar API Key de Resend (inyectada al compilar)
	resendAPIKey, err := security.DecodeSMTPPassword(ResendAPIKeyEncrypted)
	if err != nil {
		errorLogger.Printf("Advertencia: no se pudo decodificar la API Key de Resend: %v", err)
		// No hacemos panic, dejamos que intente usar la del config file si existe
	}

	// Mail Service (Resend)
	mailService := service.NewMailService(conf, resendAPIKey)
	sriService := service.NewSriService(txRepo, issuerRepo, receiptRepo, clientRepo, emissionRepo, sriClient, mailService, infoLogger)

	// ---- UI Initialization ----
	myApp := app.NewWithID("com.verith")
	myApp.Settings().SetTheme(ui.NewAppTheme())

	userUI := ui.NewUI(
		&ui.Services{
			AccService:    accService,
			CatService:    catService,
			TxService:     txService,
			UserService:   userService,
			ReportService: reportService,
			RecurService:  recurService,
			IssuerService: issuerService,
			SriService:    sriService,
			TaxService:    taxService,
		},
		infoLogger,
		errorLogger,
	)

	userUI.Init(myApp)

	// ---- App Initialization ----
	// gui.Init(a) removed as Init is called above
	infoLogger.Println("Main Window created.")

	// ---- Licensing & Startup ----
	infoLogger.Println("Checking License Status...")

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback por si falla (raro): usar carpeta local
		userConfigDir = "."
		errorLogger.Printf("Failed to get user config dir, using local: %v", err)

	}
	// 2. Definir la carpeta específica de nuestra app
	// En Windows será: AppData/Roaming/Verith/license.json
	// En Linux: .config/Verith/license.json
	licensePath := filepath.Join(userConfigDir, "Verith")
	licenseMgr := licensing.NewLicenseManager(licensePath)

	infoLogger.Println("Starting UI Loop...")
	userUI.Run(licenseMgr)
}
