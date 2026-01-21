DESKTOP_APP_SRC = ./cmd/desktop_app/main.go

.PHONY: help up down build generate dev test clean logs shell

# Default encrypted key (Placeholder for dev) if env var is not set
RESEND_ENCRYPTED_KEY ?= FAZvKFxKWlxqQVhoK1sBBAJYfydKaAxDUkJ8DgoQR3lRD3xs

run-desktop-app: ## Run the desktop application
	go build -ldflags "-X main.ResendAPIKeyEncrypted=$(RESEND_ENCRYPTED_KEY)" -tags wayland -o ./build/desktop_app $(DESKTOP_APP_SRC)
	./build/desktop_app

build-desktop-app: ## Build the desktop application
	go build -ldflags "-X main.ResendAPIKeyEncrypted=$(RESEND_ENCRYPTED_KEY)" -o ./build/desktop_app $(DESKTOP_APP_SRC)

db-up: ## Start the database container
	docker-compose up -d

db-down: ## Stop the database container
	docker-compose down

db-logs: ## View database logs
	docker-compose logs -f db

dist-windows: ## Build and package for Windows with Portable Postgres
	@echo "Building for Windows and preparing Portable Postgres..."
	@rm -rf dist/windows
	@mkdir -p dist/windows

	# 1. Descargar Postgres 16 (Portable) y VC++ Redistributable si no existen
	@if [ ! -f build/windows/pgsql.zip ]; then \
		echo "Downloading PostgreSQL binaries..."; \
		mkdir -p build/windows; \
		curl -L https://get.enterprisedb.com/postgresql/postgresql-16.1-1-windows-x64-binaries.zip -o build/windows/pgsql.zip; \
	fi
	@if [ ! -f build/windows/vc_redist.x64.exe ]; then \
		echo "Downloading Visual C++ Redistributable..."; \
		curl -L https://aka.ms/vs/17/release/vc_redist.x64.exe -o build/windows/vc_redist.x64.exe; \
	fi
	
	@cp build/windows/pgsql.zip dist/windows/pgsql.zip
	@cp build/windows/vc_redist.x64.exe dist/windows/vc_redist.x64.exe

	# 2. Package Verith using fyne tool (Icon + Metadata + API Key)
	# Usamos GOFLAGS para inyectar los ldflags a través de la herramienta fyne
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CGO_LDFLAGS="-static" \
	GOFLAGS="-ldflags=-X=main.ResendAPIKeyEncrypted=$(RESEND_ENCRYPTED_KEY)" \
	fyne package -os windows -src cmd/desktop_app

	# 3. Organizar archivos
	@mv "cmd/desktop_app/Verith.exe" dist/windows/Verith.exe
	@cp -r assets dist/windows/
	@mkdir -p dist/windows/config
	@cp config/config.yaml dist/windows/config/config.yaml
	
	# 4. Generar schema para inicialización
	@echo "Generating database schema..."
	@rm -f dist/windows/schema.sql
	@for f in migrations/2*.up.sql; do \
		cat "$$f" >> dist/windows/schema.sql; \
		echo "" >> dist/windows/schema.sql; \
	done

dist-windows-installer: dist-windows ## Generate Windows EXE Installer (Requires NSIS)
	@echo "Attempting to generate installer with NSIS..."
	@mkdir -p dist
	@if command -v makensis >/dev/null 2>&1; then \
		makensis build/windows/installer.nsi; \
		echo "Installer created: dist/Verith_Setup.exe"; \
	else \
		echo "Error: 'makensis' not found. Please install NSIS (sudo apt install nsis) to generate the setup file."; \
		echo "However, the portable version is ready in dist/windows/"; \
	fi

dist-windows-debug: ## Build debug version for Windows (With Console)
	@echo "Building DEBUG version for Windows (Console Enabled)..."
	@rm -rf dist/windows-debug
	@mkdir -p dist/windows-debug

	# Build with console enabled (no -H=windowsgui)
	CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -o dist/windows-debug/VerithDebug.exe $(DESKTOP_APP_SRC)

	@cp -r assets dist/windows-debug/
	@mkdir -p dist/windows-debug/config
	@cp config/config.yaml dist/windows-debug/config/config.yaml || cp config/config.yaml.example dist/windows-debug/config/config.yaml
	@sed -i 's/user: nelson/user: postgres/g' dist/windows-debug/config/config.yaml
	
	@echo "Debug build created in dist/windows-debug. Run VerithDebug.exe from a command prompt."
