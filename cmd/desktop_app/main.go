package main

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/nelsonmarro/accountable-holo/internal/application/report"
	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/database"
	persistence "github.com/nelsonmarro/accountable-holo/internal/infrastructure/persistence"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/storage"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ---- Infrastructure (Database) ----
	pool, err := database.Connect(ctx, conf)
	if err != nil {
		errorLogger.Fatalf("failed to connect to the database: %v", err)
	}
	defer pool.Close()
	infoLogger.Println("Connected to the database successfully")

	// ---- UI (Fyne) ----
	a := app.NewWithID("51af2ee4-c61c-4608-a3f1-d8576343af14")

	// ---- Infrastructure (Storage) ----
	storageService, err := storage.NewLocalStorageService(conf.Storage.AttachmentPath)
	if err != nil {
		errorLogger.Fatalf("failed to create storage service: %v", err)
	}

	// ---- Infrastructure (Repositories) ----
	accRepo := persistence.NewAccountRepository(pool)
	catRepo := persistence.NewCategoryRepository(pool)
	txRepo := persistence.NewTransactionRepository(pool)
	reportRepo := persistence.NewReportRepository(pool)
	userRepo := persistence.NewUserRepository(pool)

	// ---- Application (Report Generators) ----
	csvGen := report.NewCSVReportGenerator()
	pdfGen := report.NewPDFReportGenerator()

	// ---- Application (Services) ----
	accService := service.NewAccountService(accRepo)
	catService := service.NewCategoryService(catRepo)
	txService := service.NewTransactionService(txRepo, storageService, accService)
	userService := service.NewUserService(userRepo)
	reportService := service.NewReportService(reportRepo, txRepo, csvGen, pdfGen)

	// ---- UI Struct ----
	gui := ui.NewUI(&ui.Services{
		AccService:    accService,
		CatService:    catService,
		TxService:     txService,
		ReportService: reportService,
		UserService:   userService,
	}, infoLogger, errorLogger)

	// ---- App Initialization ----
	gui.Init(a)

	// ---- Run Application ----
	gui.Run()
}
