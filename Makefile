DESKTOP_APP_SRC = ./cmd/desktop_app/main.go

.PHONY: help up down build generate dev test clean logs shell

run-desktop-app: ## Run the desktop application
	go run -tags wayland $(DESKTOP_APP_SRC)

build-desktop-app: ## Build the desktop application
	go build -o ./build/desktop_app $(DESKTOP_APP_SRC)

db-up: ## Start the database container
	docker-compose up -d

db-down: ## Stop the database container
	docker-compose down

db-logs: ## View database logs
	docker-compose logs -f db
