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
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/nelsonmarro/accountable-holo/internal/ui/componets"
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
	estabCodeEntry := widget.NewEntry()
	estabCodeEntry.SetText("001")
	ptoEmiEntry := widget.NewEntry()
	ptoEmiEntry.SetText("001")

	rimpeSelect := widget.NewSelect([]string{"Ninguno", "Negocio Popular", "Emprendedor"}, nil)
	envSelect := widget.NewSelect([]string{"Pruebas", "Producción"}, nil)
	keepAccCheck := widget.NewCheck("Obligado a Llevar Contabilidad", nil)
	contribEntry := widget.NewEntry()

	if ui.currentUser.Role != domain.AdminRole {
		envSelect.Disable()
	}

	emissionInfo := container.NewVBox(
		widget.NewLabelWithStyle("Importante para Migraciones:", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabel("Si viene de otro sistema, puede usar un Punto de Emisión nuevo (ej: 002)\no configurar el secuencial inicial en el botón de abajo."),
		widget.NewForm(
			widget.NewFormItem("Cod. Establecimiento", estabCodeEntry),
			widget.NewFormItem("Punto de Emisión", ptoEmiEntry),
			widget.NewFormItem("Régimen RIMPE", rimpeSelect),
			widget.NewFormItem("Ambiente SRI", envSelect),
			widget.NewFormItem("Nro. Resolución", contribEntry),
			widget.NewFormItem("", keepAccCheck),
		),
		widget.NewButtonWithIcon("Gestionar Secuenciales / Migración", theme.ListIcon(), func() {
			dialogHandler := componets.NewEmissionPointDialog(ui.mainWindow, ui.Services.IssuerService)
			dialogHandler.Show()
		}),
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
