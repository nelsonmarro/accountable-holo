package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/nelsonmarro/accountable-holo/internal/application/report"
	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/database"
	persistence "github.com/nelsonmarro/accountable-holo/internal/infrastructure/persistence"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/storage"
	"github.com/nelsonmarro/accountable-holo/internal/licensing"
	"github.com/nelsonmarro/accountable-holo/internal/logging"
	"github.com/nelsonmarro/accountable-holo/internal/ui"
)

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
	a := app.NewWithID("51af2ee4-c61c-4608-a3f1-d8576343af14")
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

	// ---- Application (Report Generators) ----
	csvGen := report.NewCSVReportGenerator()
	pdfGen := report.NewPDFReportGenerator()

	// ---- Application (Services) ----
	accService := service.NewAccountService(accRepo)
	catService := service.NewCategoryService(catRepo)
	txService := service.NewTransactionService(txRepo, storageService, accService)
	userService := service.NewUserService(userRepo)
	reportService := service.NewReportService(reportRepo, txRepo, catRepo, csvGen, pdfGen)
	recurService := service.NewRecurringTransactionService(recurRepo, txRepo, infoLogger)

	// ---- UI Struct ----
	gui := ui.NewUI(&ui.Services{
		AccService:    accService,
		CatService:    catService,
		TxService:     txService,
		ReportService: reportService,
		UserService:   userService,
		RecurService:  recurService,
	}, infoLogger, errorLogger)

	// ---- App Initialization ----
	gui.Init(a)
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
	// En Windows será: AppData/Roaming/AccountableHolo/license.json
	// En Linux: .config/AccountableHolo/license.json
	licensePath := filepath.Join(userConfigDir, "AccountableHolo")
	licenseMgr := licensing.NewLicenseManager(licensePath)

	infoLogger.Println("Starting UI Loop...")
	gui.Run(licenseMgr)
}
