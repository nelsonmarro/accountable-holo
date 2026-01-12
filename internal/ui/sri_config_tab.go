package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
)

func (ui *UI) makeSriConfigTab() fyne.CanvasObject {
	// --- Componentes Fiscales ---
	rucEntry := widget.NewEntry()
	rucEntry.SetPlaceHolder("1790012345001")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Razón Social")

	tradeNameEntry := widget.NewEntry()
	tradeNameEntry.SetPlaceHolder("Nombre Comercial (Opcional)")

	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Dirección Matriz")

	estabAddrEntry := widget.NewEntry()
	estabAddrEntry.SetPlaceHolder("Dirección Establecimiento")

	estabCodeEntry := widget.NewEntry()
	estabCodeEntry.SetText("001")

	ptoEmiEntry := widget.NewEntry()
	ptoEmiEntry.SetText("002")

	contribEntry := widget.NewEntry()
	contribEntry.SetPlaceHolder("Nro. Resolución (Si aplica)")

	rimpeSelect := widget.NewSelect([]string{"Ninguno", "Negocio Popular", "Emprendedor"}, nil)
	rimpeSelect.SetSelected("Ninguno")

	envSelect := widget.NewSelect([]string{"Pruebas", "Producción"}, nil)
	envSelect.SetSelected("Pruebas") // Default seguro para empezar

	keepAccCheck := widget.NewCheck("Obligado a Llevar Contabilidad", nil)

	// --- Seguridad: Solo Admin cambia ambiente ---
	if ui.currentUser.Role != domain.AdminRole {
		envSelect.Disable() 
	}

	// --- Componentes SMTP (Correo) ---
	smtpServerEntry := widget.NewEntry()
	smtpServerEntry.SetPlaceHolder("smtp.gmail.com")

	smtpPortEntry := widget.NewEntry()
	smtpPortEntry.SetPlaceHolder("587")
	
	smtpUserEntry := widget.NewEntry()
	smtpUserEntry.SetPlaceHolder("tu_correo@gmail.com")

	smtpPassEntry := widget.NewPasswordEntry()
	smtpPassEntry.SetPlaceHolder("Contraseña de aplicación")

	smtpSslCheck := widget.NewCheck("Usar SSL/TLS", nil)
	smtpSslCheck.SetChecked(true)

	// --- Firma Electrónica ---
	p12Label := widget.NewLabel("No seleccionado")
	var p12Path string

	p12Btn := widget.NewButtonWithIcon("Buscar .p12", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				p12Path = reader.URI().Path()
				p12Label.SetText(filepath.Base(p12Path))
			}
		}, ui.mainWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".p12", ".pfx"}))
		fd.Show()
	})

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Contraseña del Certificado")

	// --- Logo ---
	logoLabel := widget.NewLabel("No seleccionado")
	var logoPath string

	logoBtn := widget.NewButtonWithIcon("Buscar Logo", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				logoPath = reader.URI().Path()
				logoLabel.SetText(filepath.Base(logoPath))
			}
		}, ui.mainWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	})

	// --- Cargar Datos Existentes ---
	go func() {
		ctx := context.Background()
		currentIssuer, _ := ui.Services.IssuerService.GetIssuerConfig(ctx)
		
		if currentIssuer != nil {
			fyne.Do(func() {
				rucEntry.SetText(currentIssuer.RUC)
				nameEntry.SetText(currentIssuer.BusinessName)
				tradeNameEntry.SetText(currentIssuer.TradeName)
				addressEntry.SetText(currentIssuer.MainAddress)
				estabAddrEntry.SetText(currentIssuer.EstablishmentAddress)
				estabCodeEntry.SetText(currentIssuer.EstablishmentCode)
				ptoEmiEntry.SetText(currentIssuer.EmissionPointCode)
				
				if currentIssuer.ContributionClass != "" {
					contribEntry.SetText(currentIssuer.ContributionClass)
				}
				if currentIssuer.RimpeType != "" {
					rimpeSelect.SetSelected(currentIssuer.RimpeType)
				}
				
				if currentIssuer.Environment == 1 {
					envSelect.SetSelected("Pruebas")
				} else {
					envSelect.SetSelected("Producción")
				}

				keepAccCheck.SetChecked(currentIssuer.KeepAccounting)

				// Cargar SMTP
				if currentIssuer.SMTPServer != nil {
					smtpServerEntry.SetText(*currentIssuer.SMTPServer)
				}
				if currentIssuer.SMTPPort != nil {
					smtpPortEntry.SetText(strconv.Itoa(*currentIssuer.SMTPPort))
				}
				if currentIssuer.SMTPUser != nil {
					smtpUserEntry.SetText(*currentIssuer.SMTPUser)
				}
				if currentIssuer.SMTPPassword != nil && *currentIssuer.SMTPPassword != "" {
					smtpPassEntry.SetPlaceHolder("******** (Guardada)")
				}
				smtpSslCheck.SetChecked(currentIssuer.SMTPSSL)

				if currentIssuer.SignaturePath != "" {
					p12Path = currentIssuer.SignaturePath
					p12Label.SetText(filepath.Base(p12Path))
					passwordEntry.SetPlaceHolder("******** (Guardada)")
				}
				
				if currentIssuer.LogoPath != "" {
					logoPath = currentIssuer.LogoPath
					logoLabel.SetText(filepath.Base(logoPath))
				}
			})
		}
	}()

	// --- Botón Guardar ---
	saveBtn := widget.NewButtonWithIcon("Guardar Configuración", theme.DocumentSaveIcon(), func() {
		// Validaciones básicas
		if rucEntry.Text == "" || nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("RUC y Razón Social son obligatorios"), ui.mainWindow)
			return
		}
		// Validar firma solo si es nueva config
		if p12Path == "" && p12Label.Text == "No seleccionado" {
			 dialog.ShowError(fmt.Errorf("debe seleccionar un archivo de firma electrónica"), ui.mainWindow)
			 return
		}

		envCode := 2
		if envSelect.Selected == "Pruebas" {
			envCode = 1
		}

		// Parsear puerto SMTP
		var port int
		if smtpPortEntry.Text != "" {
			p, err := strconv.Atoi(smtpPortEntry.Text)
			if err == nil {
				port = p
			}
		}
		
		server := smtpServerEntry.Text
		user := smtpUserEntry.Text
		pass := smtpPassEntry.Text

		issuer := &domain.Issuer{
			RUC:                  rucEntry.Text,
			BusinessName:         nameEntry.Text,
			TradeName:            tradeNameEntry.Text,
			MainAddress:          addressEntry.Text,
			EstablishmentAddress: estabAddrEntry.Text,
			EstablishmentCode:    estabCodeEntry.Text,
			EmissionPointCode:    ptoEmiEntry.Text,
			ContributionClass:    contribEntry.Text,
			RimpeType:            rimpeSelect.Selected,
			Environment:          envCode,
			KeepAccounting:       keepAccCheck.Checked,
			SignaturePath:        p12Path,
			LogoPath:             logoPath,
			IsActive:             true,
			SMTPServer:           &server,
			SMTPPort:             &port,
			SMTPUser:             &user,
			SMTPPassword:         &pass,
			SMTPSSL:              smtpSslCheck.Checked,
		}

		// Guardar (Password firma solo si se escribió algo nuevo)
		passToSave := passwordEntry.Text
		
		componets.HandleLongRunningOperation(ui.mainWindow, "Guardando Configuración...", func(ctx context.Context) error {
			return ui.Services.IssuerService.SaveIssuerConfig(ctx, issuer, passToSave)
		})
		
		dialog.ShowInformation("Éxito", "Configuración SRI guardada correctamente.", ui.mainWindow)
	})
	saveBtn.Importance = widget.HighImportance

	// --- Layout ---
	form := widget.NewForm(
		widget.NewFormItem("RUC", rucEntry),
		widget.NewFormItem("Razón Social", nameEntry),
		widget.NewFormItem("Nombre Comercial", tradeNameEntry),
		widget.NewFormItem("Dirección Matriz", addressEntry),
		widget.NewFormItem("Dirección Establecimiento", estabAddrEntry),
		widget.NewFormItem("Cod. Establecimiento", estabCodeEntry),
		widget.NewFormItem("Punto de Emisión", ptoEmiEntry),
		widget.NewFormItem("Contribuyente Especial", contribEntry),
		widget.NewFormItem("Obligado Contabilidad", keepAccCheck),
		widget.NewFormItem("Régimen RIMPE", rimpeSelect),
		widget.NewFormItem("Ambiente", envSelect),
		widget.NewFormItem("Firma Electrónica (.p12)", container.NewBorder(nil, nil, nil, p12Btn, p12Label)),
		widget.NewFormItem("Contraseña Firma", passwordEntry),
		widget.NewFormItem("Logo Empresa", container.NewBorder(nil, nil, nil, logoBtn, logoLabel)),
	)

	smtpForm := widget.NewForm(
		widget.NewFormItem("Servidor SMTP", smtpServerEntry),
		widget.NewFormItem("Puerto", smtpPortEntry),
		widget.NewFormItem("Usuario/Email", smtpUserEntry),
		widget.NewFormItem("Contraseña Email", smtpPassEntry),
		widget.NewFormItem("Seguridad", smtpSslCheck),
	)

	return container.NewVScroll(container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Configuración de Emisor SRI", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Configuración de Correo Electrónico (Envío de Facturas)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		smtpForm,
		widget.NewSeparator(),
		saveBtn,
	)))
}