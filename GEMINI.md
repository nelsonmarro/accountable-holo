# Accountable Holo

**Accountable Holo** is a desktop application for small business financial management. It tracks income, expenses, and accounts, offering reporting and reconciliation features.

## Project Overview

- **Type:** Desktop Application (GUI)
- **Language:** Go (Golang)
- **UI Framework:** Fyne v2
- **Database:** PostgreSQL
- **Architecture:** Clean Architecture / Layered (UI -> Application -> Domain -> Infrastructure)

## Directory Structure

- **`cmd/desktop_app`**: Entry point (`main.go`) and Fyne configuration.
- **`internal/domain`**: Core business entities (Transaction, Account, Category, etc.).
- **`internal/application`**: Business logic (Services) and use cases.
- **`internal/infrastructure`**: Database implementation (`postgres`), repositories, and file storage.
- **`internal/ui`**: Fyne UI components, screens (tabs), and dialogs.
- **`migrations`**: SQL migration files (managed by Soda/Buffalo).
- **`config`**: Configuration loading logic (`viper`).
- **`assets`**: Static assets like images and icons.

## Setup & Development

### Prerequisites
- **Go**: 1.18+
- **PostgreSQL**: Running instance.
- **Soda CLI**: For database migrations (`go install github.com/gobuffalo/pop/v6/soda@latest`).
- **Docker**: Optional, for running the database via `docker-compose`.

### Configuration
The application expects two config files:
1.  `config/config.yaml` (App settings) - Copy from `config/config.yaml.example`.
2.  `database.yml` (DB credentials) - Copy from `database.yml.example`.

### Database Management
- **Start DB (Docker):** `make db-up`
- **Run Migrations:** `soda db migrate up -e development`

### Build & Run
- **Run Locally:** `make run-desktop-app`
- **Build Binary:** `make build-desktop-app`
- **Windows Cross-Compile:** `make dist-windows` (requires MinGW)

## Architecture Details

- **Dependency Injection:** Dependencies (Repositories, Services) are manually injected in `main.go`.
- **Database Access:** Uses `pgx/v5` connection pool.
- **UI State:** Fyne's binding and widgets are used. The UI layer interacts with Services, not Repositories directly.
- **Logging:** Custom logging setup in `internal/logging` (writes to file/stdout).

## Key Files
- `cmd/desktop_app/main.go`: Application bootstrap and wiring.
- `internal/ui/fyne_ui.go`: Main UI loop and window setup.
- `internal/domain/transaction.go`: The central `Transaction` entity.
- `Makefile`: Central control for build and run tasks.
