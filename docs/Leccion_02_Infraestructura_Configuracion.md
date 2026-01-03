# Lección 2: Infraestructura con Docker y Gestión de Configuración 🐳🛠️

En la lección anterior levantamos nuestra primera ventana. Hoy vamos a sentar las bases profesionales de nuestra aplicación: una base de datos aislada y un sistema de configuración flexible.

¡Nada de instalar bases de datos manualmente ni poner contraseñas en el código!

---

## 🎯 Objetivos de la Lección

1.  Levantar **PostgreSQL** usando **Docker Compose**.
2.  Crear un archivo de configuración **YAML** para separar credenciales del código.
3.  Escribir un **Loader en Go** para leer esa configuración.

---

## 🐳 Parte 1: Base de Datos con Docker

En lugar de ensuciar nuestro sistema operativo instalando PostgreSQL, usaremos un contenedor.

### 1. Crear el archivo `docker-compose.yml`

Crea este archivo en la **raíz** de tu proyecto (`accountable-holo/docker-compose.yml`):

```yaml
version: '3.8'

services:
  db:
    image: postgres:16
    container_name: accountable_holo_db
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: accountableholodb
    ports:
      - "5432:5432" # Puerto PC : Puerto Contenedor
    volumes:
      # Guardamos los datos fuera del contenedor para no perderlos al reiniciar
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### 2. Levantar la Base de Datos

Abre tu terminal en la carpeta del proyecto y ejecuta:

```bash
docker-compose up -d
```

-   **`up`**: Crea e inicia los contenedores.
-   **`-d`**: "Detached mode" (corre en segundo plano).

> **Verificación:** Si abres **TablePlus** o cualquier cliente SQL, deberías poder conectarte a `localhost`, puerto `5432`, usuario `postgres`, contraseña `password`.

---

## ⚙️ Parte 2: El Archivo de Configuración

Vamos a separar la configuración de nuestro código fuente.

### 1. Estructura de Carpetas

Crea una nueva carpeta llamada `config` en la raíz.

```bash
mkdir config
```

### 2. El archivo `config.yaml`

Crea `config/config.yaml`. Aquí vivirán tus credenciales (¡no subas esto a producción!):

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  dbname: "accountableholodb"
  sslmode: "disable" # En local no usamos SSL
```

---

## 🧠 Parte 3: Leyendo la Configuración en Go

Go no sabe leer YAML nativamente, así que usaremos una librería estándar de la comunidad.

### 1. Instalar la librería YAML

```bash
go get gopkg.in/yaml.v3
```

### 2. El Código del Loader (`config.go`)

Crea el archivo `config/config.go`. Este código convertirá el texto del archivo YAML en una estructura (Struct) de Go que podamos usar.

```go
package config

import (
 "fmt"
 "os"
 "path/filepath"

 "gopkg.in/yaml.v3"
)

// Config representa la estructura de nuestro archivo config.yaml
// Las etiquetas `yaml:"..."` le dicen a Go qué campo buscar en el archivo.
type Config struct {
 Database struct {
  Host     string `yaml:"host"`
  Port     int    `yaml:"port"`
  User     string `yaml:"user"`
  Password string `yaml:"password"`
  DBName   string `yaml:"dbname"`
  SSLMode  string `yaml:"sslmode"`
 } `yaml:"database"`
}

// LoadConfig lee el archivo y devuelve la estructura llena o un error
func LoadConfig(path string) (*Config, error) {
 // Construimos la ruta completa al archivo
 configPath := filepath.Join(path, "config.yaml")

 // 1. Leemos los bytes del archivo
 file, err := os.ReadFile(configPath)
 if err != nil {
  return nil, fmt.Errorf("error leyendo archivo config: %w", err)
 }

 // 2. Decodificamos el YAML en nuestra estructura
 var config Config
 if err := yaml.Unmarshal(file, &config); err != nil {
  return nil, fmt.Errorf("error parseando yaml: %w", err)
 }

 return &config, nil
}
```

---

## 🔌 Parte 4: Integración en el Main

Ahora vamos a probar que todo funciona modificando nuestro punto de entrada.

Edita `cmd/desktop_app/main.go`:

```go
package main

import (
 "log"

 "fyne.io/fyne/v2/app"
 "fyne.io/fyne/v2/widget"
 
 // Importamos nuestro paquete de configuración
 // Asegúrate de cambiar "github.com/..." por TU nombre de módulo
 "github.com/TU_USUARIO/accountable-holo/config"
)

func main() {
 // 1. Cargar la configuración antes de iniciar la UI
 // Pasamos "." indicando que busque la carpeta config en el directorio actual
 conf, err := config.LoadConfig("config")
 if err != nil {
  // Si no hay config, la app no debe arrancar (Fail Fast)
  log.Fatalf("No se pudo cargar la configuración: %v", err)
 }

 log.Printf("Configuración cargada exitosamente. DB: %s", conf.Database.DBName)

 // 2. Iniciar la App (Código de la Lección 1)
 myApp := app.NewWithID("com.tu_usuario.accountable-holo")
 myWindow := myApp.NewWindow("Accountable Holo")

 // Mostramos un mensaje de éxito en la ventana
 myWindow.SetContent(widget.NewLabel("Sistema Configurado: " + conf.Database.DBName))
 
 myWindow.ShowAndRun()
}
```

---

## ✅ Ejecución y Verificación

Ejecuta nuevamente el proyecto:

```bash
go run ./cmd/desktop_app/main.go
```

**Deberías ver:**
1.  En la terminal: `Configuración cargada exitosamente. DB: accountableholodb`
2.  En la ventana: Un label que dice "Sistema Configurado: accountableholodb".

---

## 📝 Resumen

-   **Docker Compose:** Nos permite definir nuestra infraestructura como código (`infra-as-code`).
-   **YAML:** Formato legible por humanos, ideal para configuraciones.
-   **Struct Tags:** (`yaml:"host"`) Son metadatos que ayudan a las librerías de Go a mapear datos externos a estructuras internas.
-   **Fail Fast:** Si la configuración falla, detenemos el programa (`log.Fatal`) inmediatamente para evitar errores raros más adelante.

¡Nos vemos en la **Lección 3**, donde conectaremos Go a la base de datos usando **pgx**! 🚀