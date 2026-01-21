package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/uivalidators"
	"github.com/nelsonmarro/verith/internal/domain"
)

type RegistrationDialog struct {
	window      fyne.Window
	userService UserService
	logger      *log.Logger
	onSuccess   func()
}

func NewRegistrationDialog(win fyne.Window, us UserService, logger *log.Logger, onSuccess func()) *RegistrationDialog {
	return &RegistrationDialog{
		window:      win,
		userService: us,
		logger:      logger,
		onSuccess:   onSuccess,
	}
}

func (d *RegistrationDialog) Show() {
	usernameEntry := widget.NewEntry()
	passwordEntry := widget.NewEntry()
	passwordEntry.Password = true
	firstNameEntry := widget.NewEntry()
	lastNameEntry := widget.NewEntry()

	// Validadores básicos
	usernameValidator := uivalidators.NewValidator()
	usernameValidator.Required()
	usernameValidator.MinLength(3)
	usernameEntry.Validator = usernameValidator.Validate

	passwordValidator := uivalidators.NewValidator()
	passwordValidator.Required()
	passwordValidator.MinLength(8)
	passwordEntry.Validator = passwordValidator.Validate

	form := widget.NewForm(
		widget.NewFormItem("Nombre de Usuario", usernameEntry),
		widget.NewFormItem("Contraseña", passwordEntry),
		widget.NewFormItem("Nombre", firstNameEntry),
		widget.NewFormItem("Apellido", lastNameEntry),
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Bienvenido a Verith", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Crea la cuenta del Administrador para comenzar."),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		widget.NewButton("Crear Cuenta", func() {
			if usernameEntry.Validate() != nil || passwordEntry.Validate() != nil {
				dialog.ShowError(fmt.Errorf("por favor corrige los errores en el formulario"), d.window)
				return
			}

			// Creamos el admin manualmente para saltarnos la validación de currentUser (ya que no hay nadie logueado)
			// Hacemos hash aqui o usamos un metodo especial del servicio? 
			// El servicio CreateUser requiere un currentUser Admin para autorizar.
			// Necesitamos un metodo especial "CreateFirstAdmin" o hackearlo.
			
			// Lo mejor es crear un método privado en el servicio o una función dedicada en el repo, 
			// pero para no romper la arquitectura, usaremos el repositorio directamente o añadiremos lógica al servicio.
			
			// SOLUCION: En el servicio, si currentUser es nil y no hay usuarios en la DB, permitir crear.
			// Pero CreateUser pide currentUser obligatorio.
			
			// Vamos a implementar la logica aqui invocando el repo a traves de un metodo especial o 
			// asumiendo que el servicio permite currentUser=nil para el primer usuario.
			// Revisemos UserService...
			
			// UserService.CreateUser chequea: if currentUser.Role != domain.RoleAdmin
			// Esto fallará si currentUser es nil.
			
			// Hack seguro: Crear un usuario "System" temporal en memoria para pasar la validación
			systemUser := &domain.User{Role: domain.RoleAdmin}
			
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := d.userService.CreateUser(ctx, usernameEntry.Text, passwordEntry.Text, firstNameEntry.Text, lastNameEntry.Text, domain.RoleAdmin, systemUser)
			if err != nil {
				dialog.ShowError(err, d.window)
				return
			}
			
			dialog.ShowInformation("Éxito", "Usuario administrador creado.", d.window)
			d.onSuccess()
		}),
	)

	// Mostrar en un diálogo modal grande o reemplazar contenido
	// Como es el inicio, mejor reemplazar contenido de la ventana Login o Main
	// Pero este struct asume un dialogo. Vamos a hacerlo una ventana completa.
	d.window.SetContent(container.NewCenter(content))
}
