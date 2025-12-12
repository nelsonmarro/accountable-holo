package componets

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// HandleLongRunningOperation shows a cancelable progress dialog and runs the provided function in a goroutine.
// It handles error display and dialog dismissal automatically.
func HandleLongRunningOperation(win fyne.Window, title string, operation func(ctx context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())

	progressBar := widget.NewProgressBarInfinite()
	cancelBtn := widget.NewButton("Cancelar", func() {
		cancel()
	})

	progressContent := container.NewVBox(
		widget.NewLabel(title),
		progressBar,
		cancelBtn,
	)

	progressDialog := dialog.NewCustomWithoutButtons("Procesando...", progressContent, win)
	progressDialog.Show()

	go func() {
		defer fyne.Do(func() { progressDialog.Hide() })

		err := operation(ctx)
		if err != nil {
			// Don't show an error if the user cancelled the operation.
			if ctx.Err() == nil {
				fyne.Do(func() {
					dialog.ShowError(err, win)
				})
			}
		}
	}()
}
