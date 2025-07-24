package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ReportDialog struct {
	parentWindow fyne.Window
	formatSelect *widget.Select
	dialog       dialog.Dialog
	onGenerate   func(format string)
}

func NewReportDialog(parentWindow fyne.Window, onGenerate func(format string)) *ReportDialog {
	rd := &ReportDialog{
		parentWindow: parentWindow,
		onGenerate:   onGenerate,
	}

	rd.formatSelect = widget.NewSelect([]string{"PDF", "CSV"}, nil)
	rd.formatSelect.SetSelected("PDF") // Default selection

	formItems := []*widget.FormItem{
		{Text: "Formato de Reporte", Widget: rd.formatSelect},
	}

	callback := func(confirmed bool) {
		if !confirmed {
			return
		}
		rd.onGenerate(rd.formatSelect.Selected)
	}

	rd.dialog = dialog.NewForm(
		"Generar Reporte",
		"Generar",
		"Cancelar",
		formItems,
		callback,
		rd.parentWindow,
	)

	return rd
}

func (rd *ReportDialog) Show() {
	rd.dialog.Resize(fyne.NewSize(300, 200))
	rd.dialog.Show()
}
