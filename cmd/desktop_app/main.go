package main

import (
	"context"
	"log"
	"time"

	"github.com/nelsonmarro/accountable-holo/config"
	"github.com/nelsonmarro/accountable-holo/internal/application/service"
	"github.com/nelsonmarro/accountable-holo/internal/infrastructure/database"
	persistence "github.com/nelsonmarro/accountable-holo/internal/infrastructure/prersistence"
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

	// ---- Infrastructure (Repositories) ----
	accRepo := persistence.NewAccountRepository(pool)

	// ---- Application (Services) ----
	accService := service.NewAccountService(accRepo)

	// ---- UI (Fyne) ----
	gui := ui.NewUI(accService)

	gui.Run()
}
