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

dist-windows: ## Build and package for Windows
	@echo "Building for Windows using fyne tool..."
	@rm -rf dist/windows
	@mkdir -p dist/windows

	# 1. Package using fyne tool
	# We MUST keep the cross-compiler and target OS information for CGO to work
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	fyne package -os windows -icon $(shell pwd)/assets/accounts_tab_icon_dark.png -name AccountableHolo -src cmd/desktop_app

	# 2. Move the resulting executable to the dist folder
	@mv cmd/desktop_app/AccountableHolo.exe dist/windows/

	# 3. Copy Config
	@mkdir -p dist/windows/config
	@cp config/config.yaml dist/windows/config/config.yaml || cp config/config.yaml.example dist/windows/config/config.yaml
	@sed -i 's/user: nelson/user: postgres/g' dist/windows/config/config.yaml

	# 4. Generate Database Schema (requires running DB)
	@echo "Generating database schema..."
	@docker compose exec -T db pg_dump -U postgres -d accountableholodb --schema-only > dist/windows/schema.sql

	# 5. Create Windows Setup Script
	@echo "@echo off" > dist/windows/setup_db.bat
	@echo "echo Setting up Accountable Holo Database..." >> dist/windows/setup_db.bat
	@echo "set /p PGPASSWORD=Enter the password you set for the 'postgres' user during installation: " >> dist/windows/setup_db.bat
	@echo "set PGUSER=postgres" >> dist/windows/setup_db.bat
	@echo "REM Check if psql is in PATH, otherwise try default location" >> dist/windows/setup_db.bat
	@echo "where psql >nul 2>nul" >> dist/windows/setup_db.bat
	@echo "if %ERRORLEVEL% NEQ 0 set PATH=%%PATH%%;C:\Program Files\PostgreSQL\16\bin" >> dist/windows/setup_db.bat
	@echo "createdb -w accountableholodb" >> dist/windows/setup_db.bat
	@echo "psql -d accountableholodb -f schema.sql" >> dist/windows/setup_db.bat
	@echo "echo Done! You can now run AccountableHolo.exe" >> dist/windows/setup_db.bat
	@echo "pause" >> dist/windows/setup_db.bat

	@echo "Windows distribution package created in dist/windows"
