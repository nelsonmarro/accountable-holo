# Verith

**Verith** is a desktop application for small business financial management. It tracks income, expenses, and accounts, offering reporting and reconciliation features.

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

## Status Update (2026-01-26) - Feature Polish & Localization

### 1. 🌍 Localización y Formato de Fechas (Completado)
- **Estandarización:** Se ha forzado el formato `DD/MM/YYYY` en toda la aplicación.
- **LatinDateEntry:** Se implementó un widget de fecha personalizado que ignora el locale del sistema operativo, garantizando que el usuario siempre vea y escriba fechas en formato latinoamericano.
- **Traducciones:** Se inyectaron diccionarios en español para los widgets internos de Fyne (Botones OK/Cancelar, meses y días del calendario).

### 2. ⚡ Nueva Experiencia de Usuario (UX) en Transacciones
Se ha rediseñado la barra de herramientas y los flujos de registro para mayor claridad:
- **Botones Diferenciados:**
    - **Venta (Verde):** Flujo detallado para ingresos. Incluye gestión de ítems, selección de cliente (SRI) y cálculo automático de IVA.
    - **Gasto (Naranja):** Flujo simplificado para egresos. Registro rápido de monto total, descripción y adjunto, sin cálculos de impuestos (Tarifa 0% forzada).
- **Menú de Herramientas (⚙️):** Se consolidaron las funciones secundarias (Reportes, Reconciliación, Gestión de Recurrentes, Cola SRI y Recarga de Datos) en un menú desplegable para limpiar la interfaz visual.
- **Filtros Inteligentes:** Los buscadores de categorías ahora filtran automáticamente según el tipo de operación (solo verás categorías de ingreso al registrar ventas y viceversa).

### 3. ⚙️ Configuración de IVA Predeterminado (SRI)
- **Eficiencia:** Ahora puedes configurar un IVA por defecto en la pestaña de Configuración SRI (ej: IVA 15%).
- **Automatización:** Al agregar ítems a una venta, el sistema seleccionará y bloqueará automáticamente el IVA configurado para evitar errores. 
- **Flexibilidad:** Se incluyó un checkbox "Cambiar IVA manual" para permitir excepciones puntuales.
- **Cumplimiento 2026:** Se actualizaron todos los códigos de impuestos según la normativa vigente del SRI (Tarifas 15%, 13%, 5%, 0%, etc.).

### 4. 🔄 Motor de Transacciones Recurrentes
- **Detección Inteligente:** Al editar un gasto, el sistema detecta mediante heurística si pertenece a una regla de recurrencia activa y marca el estado correspondiente.
- **Ejecución Inmediata:** Si creas una regla con fecha de "Próxima Ejecución" para hoy, el sistema genera la transacción al instante y refresca la tabla general automáticamente.
- **Seguridad de Datos:** Las transacciones automáticas ahora heredan correctamente el usuario logueado y se desglosan en ítems para mantener la integridad del historial.

### 5. 🛠️ Estabilidad y Pruebas
- **Tolerancia a Fallos:** Los diálogos de transacciones ahora funcionan incluso si no se ha completado la configuración del emisor (usando valores seguros por defecto).
- **Pruebas:** Se añadieron pruebas unitarias para `ItemDialog` (lógica de IVA) y `AddExpenseDialog` (validaciones de egresos).
- **Limpieza de Código:** Se corrigieron múltiples pánicos por punteros nulos y errores de coincidencia en argumentos de base de datos.

**Nota para el Usuario:** 
Para activar el **IVA por defecto**, diríjase a `Configuración SRI > IVA Predeterminado`, seleccione su tarifa habitual y pulse `Guardar Cambios`. A partir de ese momento, el registro de ventas será mucho más rápido.

**Next Steps:**
- Final user acceptance testing.
- Distribution packaging.

