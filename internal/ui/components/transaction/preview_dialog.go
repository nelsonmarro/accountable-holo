package transaction

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PreviewDialog holds the state for the attachment preview dialog.
type PreviewDialog struct {
	mainWin fyne.Window
	fileURI fyne.URI
}

// NewPreviewDialog creates a new dialog handler for previewing attachments.
func NewPreviewDialog(win fyne.Window, fileURI fyne.URI) *PreviewDialog {
	return &PreviewDialog{
		mainWin: win,
		fileURI: fileURI,
	}
}

// Show creates and displays the dialog.
func (d *PreviewDialog) Show() {
	var content fyne.CanvasObject

	// Attempt to load the file as an image for preview
	image := canvas.NewImageFromURI(d.fileURI)
	image.FillMode = canvas.ImageFillContain

	// Check if the image was loaded successfully.
	if image.Resource == nil || image.Resource.Name() == "" {
		// It's not a previewable image, show a generic icon and label
		fileIcon := widget.NewIcon(theme.FileIcon())
		fileNameLabel := widget.NewLabel(d.fileURI.Name())
		fileNameLabel.Alignment = fyne.TextAlignCenter
		content = container.NewVBox(fileIcon, fileNameLabel)
	} else {
		// It's an image, so use it as the content
		content = image
	}

	// Create the "Save As..." button
	saveAsBtn := widget.NewButton("Save As...", d.handleSaveAs)

	// Create the main dialog content
	dialogContent := container.NewBorder(
		nil,
		container.NewCenter(saveAsBtn),
		nil,
		nil,
		content,
	)

	// Create and show the dialog
	dlg := dialog.NewCustom(d.fileURI.Name(), "Close", dialogContent, d.mainWin)
	dlg.Resize(fyne.NewSize(400, 300))
	dlg.Show()
}

// handleSaveAs is the callback for the "Save As..." button.
func (d *PreviewDialog) handleSaveAs() {
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.mainWin)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Open the source file from our app's storage
		sourceFile, err := storage.Reader(d.fileURI)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open source file reader: %w", err), d.mainWin)
			return
		}
		defer sourceFile.Close()

		// Copy the data to the destination chosen by the user
		_, err = io.Copy(writer, sourceFile)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save file: %w", err), d.mainWin)
			return
		}
	}, d.mainWin)

	fileSaveDialog.SetFileName(d.fileURI.Name())
	fileSaveDialog.Show()
}
