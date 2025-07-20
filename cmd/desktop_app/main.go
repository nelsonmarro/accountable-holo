package main

import (
	"context"
	"log"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/database"
	persistence "github.com/nelsonmarro/accountable-holo/internal/infrastructure/persistence"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/storage"
	"github.com/nelsonmarro/accountable-holo/internal/ui"
)

func main() {
	conf := config.LoadConfig("config")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ---- Infrastructure (Database) ----
	pool, err := database.Connect(ctx, conf)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to the database successfully")

	// ---- UI (Fyne) ----
	// 1. Create the Fyne App first.
	a := app.NewWithID("51af2ee4-c61c-4608-a3f1-d8576343af14")

	// ---- Infrastructure (Storage) ----
	storageService, err := storage.NewLocalStorageService(conf.Storage.AttachmentPath)
	if err != nil {
		log.Fatalf("failed to create storage service: %v", err)
	}

	// ---- Infrastructure (Repositories) ----
	accRepo := persistence.NewAccountRepository(pool)
	catRepo := persistence.NewCategoryRepository(pool)
	txRepo := persistence.NewTransactionRepository(pool)

	// ---- Application (Services) ----
	accService := service.NewAccountService(accRepo)
	catService := service.NewCategoryService(catRepo)
	txService := service.NewTransactionService(txRepo, storageService)

	// 2. Create UI struct.
	gui := ui.NewUI(&ui.Services{
		AccService: accService,
		CatService: catService,
		TxService:  txService,
	})

	// 3. Initialize the UI with the app object.
	gui.Init(a)

	// 4. Run the application.
	gui.Run()
}
