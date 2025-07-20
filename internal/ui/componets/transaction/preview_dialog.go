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
	"fyne.io/fyne/v2/widget"
)

// PreviewDialog holds the state for the attachment preview dialog.
type PreviewDialog struct {
	mainWin      fyne.Window
	storagePath  string // The path of the file within our app's storage
	originalName string // The original name of the file for saving
}

// NewPreviewDialog creates a new dialog handler for previewing attachments.
func NewPreviewDialog(win fyne.Window, storagePath string) *PreviewDialog {
	return &PreviewDialog{
		mainWin:      win,
		storagePath:  storagePath,
		originalName: filepath.Base(storagePath), // Extract filename from path
	}
}

// Show creates and displays the dialog.
func (d *PreviewDialog) Show() {
	var content fyne.CanvasObject

	// Attempt to load the file as an image for preview
	image := canvas.NewImageFromFile(d.storagePath)
	image.FillMode = canvas.ImageFillContain

	// Use a more reliable check on the image's resource.
	// if image.Resource == nil || image.Resource.Name() == "" {
	// 	// It's not a previewable image, show a generic icon and label
	// 	fmt.Printf("File %s is not a valid image, showing generic icon\n", d.storagePath)
	// 	fileIcon := widget.NewIcon(theme.FileIcon())
	// 	fileNameLabel := widget.NewLabel(d.originalName)
	// 	fileNameLabel.Alignment = fyne.TextAlignCenter
	// 	content = container.NewVBox(fileIcon, fileNameLabel)
	// } else {
	// 	// It's an image, so use it as the content
	content = image
	// }

	// Create the "Save As..." button
	saveAsBtn := widget.NewButton("Save As...", d.handleSaveAs)

	// Create the main dialog content
	dialogContent := container.NewBorder(
		nil,
		container.NewCenter(saveAsBtn), // Center the button at the bottom
		nil,
		nil,
		content, // The image or the icon/label
	)

	// Create and show the dialog
	dlg := dialog.NewCustom(d.originalName, "Close", dialogContent, d.mainWin)
	dlg.Resize(fyne.NewSize(400, 300)) // Give it a reasonable default size
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
			// User cancelled
			return
		}
		defer writer.Close()

		// Open the source file from our app's storage
		sourceFile, err := os.Open(d.storagePath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open source file: %w", err), d.mainWin)
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

	// Suggest the original filename to the user
	fileSaveDialog.SetFileName(d.originalName)
	fileSaveDialog.Show()
}
