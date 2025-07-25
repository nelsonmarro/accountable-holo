package componets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ReportDialog struct {
	parentWindow fyne.Window
	formatSelect *widget.Select
	dialog       dialog.Dialog
	onGenerate   func(format string, outputPath string)
}

func NewReportDialog(parentWindow fyne.Window, onGenerate func(format string,
	outputPath string),
) *ReportDialog {
	rd := &ReportDialog{
		parentWindow: parentWindow,
		onGenerate:   onGenerate,
	}
	return rd
}

func (rd *ReportDialog) Show() {
	rd.formatSelect = widget.NewSelect([]string{"PDF", "CSV"}, nil)
	rd.formatSelect.SetSelected("PDF") // Default selection

	formItems := []*widget.FormItem{
		{Text: "Formato de Reporte", Widget: rd.formatSelect},
	}

	callback := func(confirmed bool) {
		if !confirmed {
			return
		}
		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, rd.parentWindow)
				return
			}

			if writer == nil { // User canceled the save dialog}
				return
			}
			defer writer.Close()

			rd.onGenerate(rd.formatSelect.Selected, writer.URI().Path())
		}, rd.parentWindow)

		fileSaveDialog.SetFileName("reporte." + strings.ToLower(rd.formatSelect.Selected))
		fileSaveDialog.Show()
	}

	rd.dialog = dialog.NewForm(
		"Generar Reporte",
		"Generar",
		"Cancelar",
		formItems,
		callback,
		rd.parentWindow,
	)
	rd.dialog.Resize(fyne.NewSize(300, 200))
	rd.dialog.Show()
}
