package transaction

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPreviewDialog(t *testing.T) {
	// Setup a test Fyne app and window
	a := test.NewApp()
	w := a.NewWindow("Test")
	defer w.Close()

	// Create a temporary file to act as our attachment
	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "test_attachment.txt")
	err := os.WriteFile(tempFilePath, []byte("test data"), 0666)
	require.NoError(t, err)

	// Create the dialog
	previewDialog := NewPreviewDialog(w, tempFilePath)

	// Assertions
	assert.NotNil(t, previewDialog, "PreviewDialog should not be nil")
	assert.Equal(t, w, previewDialog.mainWin, "Window should be set correctly")
	assert.Equal(t, tempFilePath, previewDialog.storagePath, "Storage path should be set correctly")
	assert.Equal(t, "test_attachment.txt", previewDialog.originalName, "Original name should be extracted correctly")
}

func TestPreviewDialog_Show(t *testing.T) {
	// This test primarily ensures that the Show() method can be called without crashing.
	// It's difficult to test the visual output in a unit test.

	t.Run("with non-image file", func(t *testing.T) {
		a := test.NewApp()
		w := a.NewWindow("Test")
		defer w.Close()

		tempDir := t.TempDir()
		tempFilePath := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(tempFilePath, []byte("test"), 0666)
		require.NoError(t, err)

		previewDialog := NewPreviewDialog(w, tempFilePath)

		// We can't easily test the dialog is visible, but we can ensure no panics occur.
		assert.NotPanics(t, func() {
			previewDialog.Show()
		})
	})

	t.Run("with a (non-existent) image file", func(t *testing.T) {
		// Fyne's canvas.NewImageFromFile doesn't error on non-existent files,
		// so this tests the fallback logic.
		a := test.NewApp()
		w := a.NewWindow("Test")
		defer w.Close()

		previewDialog := NewPreviewDialog(w, "non_existent_image.png")

		assert.NotPanics(t, func() {
			previewDialog.Show()
		})
	})
}
