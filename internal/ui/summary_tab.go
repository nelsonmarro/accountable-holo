package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) makeSummaryTab() fyne.CanvasObject {
	// 1. Filtros (Barra Superior Compacta)
	filterBar := ui.makeFilterBar()

	// 2. Métricas (KPIs)
	// Ya están inicializadas en makeSummaryCard, pero reusaremos los objetos de texto
	// para un layout más limpio sin tarjetas anidadas innecesarias.
	kpiContainer := ui.makeKPIGrid()

	// 3. Gráficos (Dos columnas)
	ui.summaryChartsContainer = container.NewGridWithColumns(2)
	// Inicialmente vacío, se llena en generateSummary

	// 4. Presupuestos
	ui.summaryBudgetContainer = container.NewVBox()
	budgetCard := widget.NewCard("Control de Presupuestos", "", ui.summaryBudgetContainer)

	// --- Layout Principal ---
	
	// Contenedor Central (Gráficos + Presupuestos)
	// Usamos un Split o simplemente VBox scrolleable
	
	dashboardContent := container.NewVBox(
		container.NewPadded(filterBar),
		container.NewPadded(kpiContainer),
		container.NewPadded(ui.summaryChartsContainer),
		container.NewPadded(budgetCard),
	)

	return container.NewVScroll(dashboardContent)
}

func (ui *UI) makeKPIGrid() fyne.CanvasObject {
	// Asegurar que los widgets existen (se crean en makeSummaryCard logicamente, 
	// pero aquí los instanciamos si es necesario o reusamos la lógica de inicialización)
	if ui.summaryTotalIncome == nil {
		ui.makeSummaryCard() // Hack para inicializar los punteros
	}

	card1 := widget.NewCard("Ingresos", "", container.NewCenter(ui.summaryTotalIncome))
	card2 := widget.NewCard("Egresos", "", container.NewCenter(ui.summaryTotalExpenses))
	card3 := widget.NewCard("Neto", "", container.NewCenter(ui.summaryNetProfitLoss))

	return container.NewGridWithColumns(3, card1, card2, card3)
}

// makeFilterBar crea una versión horizontal y compacta de los filtros

func (ui *UI) makeFilterBar() fyne.CanvasObject {

	ui.makeFilterCard() 



	rangeSelect := ui.summaryDateRangeSelect

	accountSelect := ui.summaryAccountSelect

	

	// Forzar un ancho mínimo para que el texto no se corte

	// En Fyne, una forma limpia es usar un contenedor con tamaño mínimo

	rangeContainer := container.NewStack(rangeSelect)

	accountContainer := container.NewStack(accountSelect)



		// Fechas personalizadas con más espacio



		startDate := ui.summaryStartDateEntry



		endDate := ui.summaryEndDateEntry



	



		// Envolver en contenedores con ancho mínimo



		startContainer := container.NewStack(startDate)



		endContainer := container.NewStack(endDate)



		



		// Crear un layout que fuerce el tamaño



		dateWrapper := container.NewGridWithColumns(2,



			container.NewVBox(widget.NewLabel("Desde:"), startContainer),



			container.NewVBox(widget.NewLabel("Hasta:"), endContainer),



		)



	



		customDatesContainer := container.NewPadded(dateWrapper)



		customDatesContainer.Hide()



	



	rangeSelect.OnChanged = func(s string) {

		if s == "Personalizado" {

			customDatesContainer.Show()

		} else {

			customDatesContainer.Hide()

		}

	}



	generateBtn := widget.NewButtonWithIcon("Actualizar Dashboard", theme.ViewRefreshIcon(), func() {

		go ui.generateSummary()

	})

	generateBtn.Importance = widget.HighImportance



	// Rejilla para que los elementos crezcan proporcionalmente y no se amontonen

	filterGrid := container.NewGridWithColumns(3,

		container.NewVBox(widget.NewLabelWithStyle("Rango de Tiempo", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), rangeContainer),

		container.NewVBox(widget.NewLabelWithStyle("Cuenta Bancaria", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), accountContainer),

		container.NewVBox(widget.NewLabel(" "), generateBtn), // Label vacío para alinear el botón

	)



	return container.NewVBox(

		filterGrid,

		customDatesContainer,

	)

}
