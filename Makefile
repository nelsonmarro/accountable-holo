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
	@echo "Building for Windows using fyne tool and metadata..."
	@rm -rf dist/windows
	@mkdir -p dist/windows

	# 1. Package using fyne tool (uses FyneApp.toml)
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CGO_LDFLAGS="-static" \
	fyne package -os windows -src cmd/desktop_app

	# 2. Move the resulting executable to the dist folder
	@mv "cmd/desktop_app/Accountable Holo.exe" dist/windows/AccountableHolo.exe

	# 3. Copy Assets and Config
	@cp -r assets dist/windows/
	@mkdir -p dist/windows/config
	@cp config/config.yaml dist/windows/config/config.yaml || cp config/config.yaml.example dist/windows/config/config.yaml
	@sed -i 's/user: nelson/user: postgres/g' dist/windows/config/config.yaml

	# 4. Generate Database Schema and Seed Data from migrations
	@echo "Generating clean database schema from migrations..."
	@rm -f dist/windows/schema.sql
	@for f in migrations/2*.up.sql; do \
		cat "$$f" >> dist/windows/schema.sql; \
		echo "" >> dist/windows/schema.sql; \
	done

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

dist-windows-debug: ## Build debug version for Windows (With Console)
	@echo "Building DEBUG version for Windows (Console Enabled)..."
	@rm -rf dist/windows-debug
	@mkdir -p dist/windows-debug

	# Build with console enabled (no -H=windowsgui)
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -o dist/windows-debug/AccountableHoloDebug.exe $(DESKTOP_APP_SRC)

	@cp -r assets dist/windows-debug/
	@mkdir -p dist/windows-debug/config
	@cp config/config.yaml dist/windows-debug/config/config.yaml || cp config/config.yaml.example dist/windows-debug/config/config.yaml
	@sed -i 's/user: nelson/user: postgres/g' dist/windows-debug/config/config.yaml
	
	@echo "Debug build created in dist/windows-debug. Run AccountableHoloDebug.exe from a command prompt."
