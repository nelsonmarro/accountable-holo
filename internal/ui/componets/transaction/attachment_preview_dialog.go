package transaction

import (
	"path/filepath"

	"fyne.io/fyne/v2"
)

// PreviewDialog holds the state for the attachment preview dialog.
type PreviewDialog struct {
	mainWin      fyne.Window
	storagePath  string // Path to the storage directory
	originalName string // Original name of the file
}

func NewPreviewDialog(win fyne.Window, storagePath string) *PreviewDialog {
	return &PreviewDialog{
		mainWin:      win,
		storagePath:  storagePath,
		originalName: filepath.Base(storagePath), // Extract filename from path
	}
}
