package ui

import (
	"context"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/nelsonmarro/verith/internal/ui/componets"
)

func (ui *UI) makeSriConfigTab() fyne.CanvasObject {
	// --- SECCIÓN 1: DATOS LEGALES ---
	rucEntry := widget.NewEntry()
	rucEntry.SetPlaceHolder("Ej: 1790012345001")
	nameEntry := widget.NewEntry()
	tradeNameEntry := widget.NewEntry()
	addressEntry := widget.NewEntry()
	estabAddrEntry := widget.NewEntry()

	legalForm := widget.NewForm(
		widget.NewFormItem("RUC", rucEntry),
		widget.NewFormItem("Razón Social", nameEntry),
		widget.NewFormItem("Nombre Comercial", tradeNameEntry),
		widget.NewFormItem("Dir. Matriz", addressEntry),
		widget.NewFormItem("Dir. Establecimiento", estabAddrEntry),
	)

	legalCard := widget.NewCard("Información Legal y Matriz", "Datos tal como constan en su RUC", legalForm)

	// --- SECCIÓN 2: CONFIGURACIÓN DE EMISIÓN ---
	taxOptions := []string{"Ninguno (Manual)", "IVA 15%", "IVA 13%", "IVA 8%", "IVA 5%", "IVA 0%", "No Objeto (6)", "Exento (7)"}
	defaultTaxSelect := widget.NewSelect(taxOptions, nil)
	defaultTaxSelect.SetSelected("Ninguno (Manual)")

	estabCodeEntry := widget.NewEntry()
	estabCodeEntry.SetText("001")
	ptoEmiEntry := widget.NewEntry()
	ptoEmiEntry.SetText("001")

	rimpeSelect := widget.NewSelect([]string{"Ninguno", "Negocio Popular", "Emprendedor"}, nil)
	envSelect := widget.NewSelect([]string{"Pruebas", "Producción"}, nil)
	keepAccCheck := widget.NewCheck("Obligado a Llevar Contabilidad", nil)
	contribEntry := widget.NewEntry()

	if ui.currentUser.Role != domain.RoleAdmin {
		envSelect.Disable()
	}

	migrationBtn := widget.NewButtonWithIcon("MIGRAR / AJUSTAR SECUENCIALES", theme.WarningIcon(), func() {
		dialogHandler := componets.NewEmissionPointDialog(ui.mainWindow, ui.Services.IssuerService)
		dialogHandler.Show()
	})
	migrationBtn.Importance = widget.WarningImportance

	emissionInfo := container.NewVBox(
		widget.NewRichText(&widget.TextSegment{
			Text:  "⚙️ Gestión de Secuenciales y Migración",
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true, Italic: true}},
		}),
		widget.NewLabel("IMPORTANTE: Si migra de otro software, haga clic en el botón de abajo para configurar\nel último secuencial emitido y garantizar la continuidad legal."),
		migrationBtn,
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("Cod. Establecimiento", estabCodeEntry),
			widget.NewFormItem("Punto de Emisión", ptoEmiEntry),
			widget.NewFormItem("Régimen RIMPE", rimpeSelect),
			widget.NewFormItem("Ambiente SRI", envSelect),
			widget.NewFormItem("IVA Predeterminado", defaultTaxSelect),
			widget.NewFormItem("Nro. Resolución", contribEntry),
			widget.NewFormItem("", keepAccCheck),
		),
	)

	emissionCard := widget.NewCard("Configuración de Facturación", "Defina cómo se generarán sus comprobantes", emissionInfo)

	// --- SECCIÓN 3: FIRMA Y SEGURIDAD ---
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
	passwordEntry.SetPlaceHolder("Contraseña del certificado")

	securityForm := widget.NewForm(
		widget.NewFormItem("Archivo Firma", container.NewBorder(nil, nil, nil, p12Btn, p12Label)),
		widget.NewFormItem("Contraseña", passwordEntry),
	)
	securityCard := widget.NewCard("Firma Electrónica", "Certificado digital requerido para validez legal", securityForm)

	// --- SECCIÓN 4: IDENTIDAD VISUAL ---
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

	logoContainer := container.NewVBox(
		widget.NewLabel("El logo aparecerá en el PDF (RIDE) enviado a sus clientes."),
		container.NewBorder(nil, nil, nil, logoBtn, logoLabel),
	)
	logoCard := widget.NewCard("Identidad Visual", "Personalización de comprobantes", logoContainer)

	// --- CARGAR DATOS EXISTENTES ---
	go func() {
		ctx := context.Background()
		currentIssuer, _ := ui.Services.IssuerService.GetIssuerConfig(ctx)
		if currentIssuer != nil {
			fyne.Do(func() {
				switch currentIssuer.DefaultTaxRate {
				case 4:
					defaultTaxSelect.SetSelected("IVA 15%")
				case 10:
					defaultTaxSelect.SetSelected("IVA 13%")
				case 8:
					defaultTaxSelect.SetSelected("IVA 8%")
				case 5:
					defaultTaxSelect.SetSelected("IVA 5%")
				case 0:
					defaultTaxSelect.SetSelected("IVA 0%")
				case 6:
					defaultTaxSelect.SetSelected("No Objeto (6)")
				case 7:
					defaultTaxSelect.SetSelected("Exento (7)")
				default:
					defaultTaxSelect.SetSelected("Ninguno (Manual)")
				}

				rucEntry.SetText(currentIssuer.RUC)
				nameEntry.SetText(currentIssuer.BusinessName)
				tradeNameEntry.SetText(currentIssuer.TradeName)
				addressEntry.SetText(currentIssuer.MainAddress)
				estabAddrEntry.SetText(currentIssuer.EstablishmentAddress)
				estabCodeEntry.SetText(currentIssuer.EstablishmentCode)
				ptoEmiEntry.SetText(currentIssuer.EmissionPointCode)
				contribEntry.SetText(currentIssuer.ContributionClass)
				rimpeSelect.SetSelected(currentIssuer.RimpeType)
				if currentIssuer.Environment == 1 {
					envSelect.SetSelected("Pruebas")
				} else {
					envSelect.SetSelected("Producción")
				}
				keepAccCheck.SetChecked(currentIssuer.KeepAccounting)
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

	// --- BOTÓN GUARDAR ---
	saveBtn := widget.NewButtonWithIcon("Guardar Cambios", theme.DocumentSaveIcon(), func() {
		if rucEntry.Text == "" || nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("RUC y Razón Social son obligatorios"), ui.mainWindow)
			return
		}
		if p12Path == "" && p12Label.Text == "No seleccionado" {
			dialog.ShowError(fmt.Errorf("debe seleccionar un archivo de firma electrónica"), ui.mainWindow)
			return
		}

		envCode := 2
		if envSelect.Selected == "Pruebas" {
			envCode = 1
		}

		taxRate := -1
		switch defaultTaxSelect.Selected {
		case "IVA 15%":
			taxRate = 4
		case "IVA 13%":
			taxRate = 10
		case "IVA 8%":
			taxRate = 8
		case "IVA 5%":
			taxRate = 5
		case "IVA 0%":
			taxRate = 0
		case "No Objeto (6)":
			taxRate = 6
		case "Exento (7)":
			taxRate = 7
		}

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
			DefaultTaxRate:       taxRate,
			IsActive:             true,
		}

		componets.HandleLongRunningOperation(ui.mainWindow, "Guardando Configuración...", func(ctx context.Context) error {
			return ui.Services.IssuerService.SaveIssuerConfig(ctx, issuer, passwordEntry.Text)
		}, func() {
			dialog.ShowInformation("Éxito", "Configuración actualizada correctamente.", ui.mainWindow)
		})
	})
	saveBtn.Importance = widget.HighImportance

	// --- LAYOUT FINAL ---
	content := container.NewVBox(
		legalCard,
		emissionCard,
		securityCard,
		logoCard,
		container.NewPadded(saveBtn),
	)

	return container.NewVScroll(container.NewPadded(content))
}
