package componets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type ReportDialog struct {
	parentWindow fyne.Window
	dialog       dialog.Dialog

	// Transaction Report Tab
	transactionReportFormatSelect *widget.Select
	onGenerateTransactionReport   func(format string, outputPath string)

	// Daily Report Tab
	dailyReportFormatSelect *widget.Select
	onGenerateDailyReport   func(format string, outputPath string)
}

func NewReportDialog(
	parentWindow fyne.Window,
	onGenerateTransactionReport func(format string, outputPath string),
	onGenerateDailyReport func(format string, outputPath string),
) *ReportDialog {
	return &ReportDialog{
		parentWindow:                parentWindow,
		onGenerateTransactionReport: onGenerateTransactionReport,
		onGenerateDailyReport:       onGenerateDailyReport,
	}
}

func (rd *ReportDialog) Show() {
	tabs := container.NewAppTabs(
		container.NewTabItem("Reporte Financiero Diario", rd.createDailyReportTab()),
		container.NewTabItem("Reporte de Transacciones", rd.createTransactionReportTab()),
	)

	rd.dialog = dialog.NewCustom("Generar Reporte", "Cerrar", tabs, rd.parentWindow)
	rd.dialog.Resize(fyne.NewSize(600, 230))
	rd.dialog.Show()
}

func (rd *ReportDialog) createTransactionReportTab() fyne.CanvasObject {
	rd.transactionReportFormatSelect = widget.NewSelect([]string{"PDF", "CSV"}, nil)
	rd.transactionReportFormatSelect.SetSelected("PDF")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Formato", Widget: rd.transactionReportFormatSelect},
		},
	}

	generateBtn := widget.NewButton("Generar", func() {
		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, rd.parentWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			// Fire and forget. The caller is responsible for async execution and error handling.
			rd.onGenerateTransactionReport(rd.transactionReportFormatSelect.Selected, writer.URI().Path())
		}, rd.parentWindow)
		fileSaveDialog.SetFileName("reporte_transacciones." + strings.ToLower(rd.transactionReportFormatSelect.Selected))
		fileSaveDialog.Show()
	})
	generateBtn.Importance = widget.SuccessImportance

	return container.NewVBox(form, generateBtn)
}

func (rd *ReportDialog) createDailyReportTab() fyne.CanvasObject {
	rd.dailyReportFormatSelect = widget.NewSelect([]string{"PDF", "CSV"}, nil)
	rd.dailyReportFormatSelect.SetSelected("PDF")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Formato", Widget: rd.dailyReportFormatSelect},
		},
	}

	generateBtn := widget.NewButton("Generar", func() {
		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, rd.parentWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			// Fire and forget. The caller is responsible for async execution and error handling.
			rd.onGenerateDailyReport(rd.dailyReportFormatSelect.Selected, writer.URI().Path())
		}, rd.parentWindow)
		fileSaveDialog.SetFileName("reporte_diario." + strings.ToLower(rd.dailyReportFormatSelect.Selected))
		fileSaveDialog.Show()
	})
	generateBtn.Importance = widget.SuccessImportance

	return container.NewVBox(form, generateBtn)
}