package transaction

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PreviewDialog holds the state for the attachment preview dialog.
type PreviewDialog struct {
	mainWin      fyne.Window
	storagePath  string // Path to the storage directory
	originalName string // Original name of the file
}

// NewPreviewDialog creates a new instance of PreviewDialog with the given window and storage path.
func NewPreviewDialog(win fyne.Window, storagePath string) *PreviewDialog {
	return &PreviewDialog{
		mainWin:      win,
		storagePath:  storagePath,
		originalName: filepath.Base(storagePath), // Extract filename from path
	}
}

// Show creates and displays the dialog
func (d *PreviewDialog) Show() {
	var content fyne.CanvasObject

	// Attempt to load the file as an image for preview
	image := canvas.NewImageFromFile(d.storagePath)
	image.FillMode = canvas.ImageFillContain

	// Check if the image is valid
	if image.Image == nil || image.File == "" {
		// it's not a previewable image, so wee create a generic icon and label
		fileIcon := widget.NewIcon(theme.FileIcon())
		fileNameLabel := widget.NewLabel(d.originalName)
		fileNameLabel.Alignment = fyne.TextAlignCenter
		content = container.NewVBox(fileIcon, fileNameLabel)
	} else { // it's a valid image, so we use it as the content
		content = image
	}

	// Create the "Guardar como" button
	saveAsBtn := widget.NewButton("Guardar como", d.handleSaveAs)

	dialogContent := container.NewBorder(
		nil,
		container.NewCenter(saveAsBtn),
		nil, nil,
		content,
	)

	// Create and show the dialog
	dlg := dialog.NewCustom(d.originalName, "Cerrar", dialogContent, d.mainWin)
	dlg.Resize(fyne.NewSize(560, 430))
	dlg.Show()
}

// handleSaveAs is the callback for the "Guardar como" button.
func (d *PreviewDialog) handleSaveAs() {
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.mainWin)
			return
		}

		if writer == nil { // User cancelled the save dialog}
			return
		}

		defer writer.Close()

		// Open the source file from our app's storage
		sourceFile, err := os.Open(d.storagePath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error al abrir el archivo fuente: %w", err), d.mainWin)
			return
		}
		defer sourceFile.Close()

		// Copy the content to the new file
		_, err = io.Copy(writer, sourceFile)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error al guardar el archivo: %w", err), d.mainWin)
			return
		}
	}, d.mainWin)

	// Suggest a filename based on the original file name
	fileSaveDialog.SetFileName(d.originalName)
	fileSaveDialog.Show()
}
