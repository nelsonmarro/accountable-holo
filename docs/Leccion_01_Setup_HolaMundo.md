# Lección 1: Configuración del Entorno y "Hola Mundo" con Go y Fyne 🚀

¡Bienvenidos al curso! En esta primera lección, vamos a preparar nuestro entorno de desarrollo profesional y construiremos nuestra primera ventana de escritorio.

---

## 🛠️ Herramientas Necesarias (Prerrequisitos)

Sigue las instrucciones según tu sistema operativo.

### 1. Git (Control de Versiones)

- **Windows:** Descarga e instala desde [git-scm.com](https://git-scm.com/download/win).
- **Linux (Ubuntu/Debian):**

  ```bash
  sudo apt-get update
  sudo apt-get install git
  ```

### 2. Go y Fyne (Lenguaje y Compilador de C)

Fyne requiere **Go (mínimo 1.19)**, un **compilador de C** para conectar con los drivers de gráficos y los **drivers del sistema**.

#### 🪟 Windows (Vía MSYS2)

Es la forma recomendada para evitar errores de compilación con CGO.

1. Instala **MSYS2** desde [msys2.org](https://www.msys2.org/).
2. Al finalizar, busca en el menú de inicio **"MSYS2 MinGW 64-bit"** y ábrelo.
3. Ejecuta los siguientes comandos (elige "all" si se te pregunta):

    ```bash
    pacman -Syu
    pacman -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go
    ```

4. Configura el **PATH** en MSYS2:

    ```bash
    echo "export PATH=\$PATH:~/Go/bin" >> ~/.bashrc
    ```

5. **Variables de Entorno de Windows:** Para usar otros terminales (PowerShell/CMD/VS Code), ve al "Panel de Control" -> "Editar las variables de entorno del sistema" -> "Variables de entorno" -> Busca `Path` -> Agrega: `C:\msys64\mingw64\bin`.

#### 🐧 Linux (Ubuntu/Debian)

Instala Go, GCC y las librerías de desarrollo de X11/Mesa:

```bash
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

### 3. Docker (Base de Datos) 🐳

Usaremos Docker para ejecutar PostgreSQL sin necesidad de instalaciones complejas en el sistema local.

- **Windows/Mac:** Instala [Docker Desktop](https://www.docker.com/products/docker-desktop/).
- **Linux:** Instala [Docker Engine](https://docs.docker.com/engine/install/ubuntu/).

### 4. TablePlus (Visualizador de DB) 👁️

Herramienta recomendada para explorar los datos de forma visual.

- Descarga la versión gratuita en [tableplus.com](https://tableplus.com/).

---

## 🚀 Inicialización del Proyecto

### 1. Crear el Módulo

Abre tu terminal y ejecuta:

```bash
mkdir accountable-holo
cd accountable-holo
go mod init github.com/TU_USUARIO/accountable-holo
```

### 2. Estructura de Carpetas (Clean Architecture)

Organizaremos nuestro código de forma profesional desde el inicio:

```bash
mkdir -p cmd/desktop_app
mkdir -p internal/ui
mkdir assets
```

- `cmd/`: Puntos de entrada de la aplicación.
- `internal/`: Lógica privada del negocio y UI.
- `assets/`: Iconos, imágenes y recursos estáticos.

### 3. Instalar Fyne y sus herramientas

```bash
go get fyne.io/fyne/v2@latest
go install fyne.io/tools/cmd/fyne@latest
```

---

## 💻 Código: Nuestra Primera Ventana

Crea el archivo `cmd/desktop_app/main.go` y pega el siguiente código:

```go
package main

import (
 "fyne.io/fyne/v2/app"
 "fyne.io/fyne/v2/widget"
)

func main() {
 // 1. Crear la aplicación con un ID único
 myApp := app.NewWithID("com.nombre.accountable-holo")

 // 2. Crear una nueva ventana
 myWindow := myApp.NewWindow("Accountable Holo")

 // 3. Agregar contenido (un Label simple)
 myWindow.SetContent(widget.NewLabel("¡Bienvenido a Accountable Holo!"))

 // 4. Mostrar y ejecutar la aplicación
 myWindow.ShowAndRun()
}
```

---

## ▶️ Ejecución

Para correr el proyecto, ejecuta desde la raíz:

```bash
go run ./cmd/desktop_app/main.go
```

---

## 📝 Conceptos Clave de la Lección

- **Go Modules:** Sistema de gestión de dependencias oficial de Go.
- **CGO:** Puente que permite a Go llamar código escrito en C (necesario para interactuar con la tarjeta gráfica/OpenGL).
- **Main Package:** El paquete `main` y la función `main()` son el punto de partida obligatorio para cualquier ejecutable en Go.
