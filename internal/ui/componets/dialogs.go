package componets

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// HandleLongRunningOperation shows a cancelable progress dialog and runs the provided function in a goroutine.
// onSuccess is an optional callback that runs on the UI thread after the operation completes successfully and the dialog closes.
func HandleLongRunningOperation(win fyne.Window, title string, operation func(ctx context.Context) error, onSuccess func()) {
	ctx, cancel := context.WithCancel(context.Background())

	progressBar := widget.NewProgressBarInfinite()

	var progressDialog dialog.Dialog

	cancelBtn := widget.NewButton("Cancelar", func() {
		cancel()
		if progressDialog != nil {
			progressDialog.Hide()
		}
	})

	progressContent := container.NewVBox(
		widget.NewLabel(title),
		progressBar,
		cancelBtn,
	)

	progressDialog = dialog.NewCustomWithoutButtons("Procesando...", progressContent, win)
	progressDialog.Show()

	go func() {
		err := operation(ctx)

		// Cerrar el diálogo de progreso explícitamente después de la operación
		fyne.Do(func() {
			progressDialog.Hide()
			if err == nil && onSuccess != nil {
				onSuccess()
			}
		})

		if err != nil {
			// No mostrar error si la operación fue cancelada por el usuario
			if ctx.Err() == nil {
				fyne.Do(func() {
					dialog.ShowError(err, win)
				})
			}
		}
	}()
}
