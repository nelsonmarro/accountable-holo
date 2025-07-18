package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStorage(t *testing.T) (service *LocalStorageService, cleanup func()) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	require.NoError(t, err)

	// Override the user config dir for this test
	originalUserConfigDir := os.UserConfigDir
	os.UserConfigDir = func() (string, error) {
		return tempDir, nil
	}

	service, err = NewLocalStorageService()
	require.NoError(t, err)

	// The cleanup function to be called by the test
	cleanup = func() {
		os.UserConfigDir = originalUserConfigDir
		os.RemoveAll(tempDir)
	}

	return service, cleanup
}

func TestLocalStorageService_Save(t *testing.T) {
	service, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create a dummy source file
	sourceContent := "test content"
	sourceFile, err := os.CreateTemp("", "source-*.txt")
	require.NoError(t, err)
	defer os.Remove(sourceFile.Name())
	_, err = sourceFile.WriteString(sourceContent)
	require.NoError(t, err)
	sourceFile.Close()

	destinationName := "my-test-file.txt"
	storagePath, err := service.Save(context.Background(), sourceFile.Name(), destinationName)
	require.NoError(t, err)

	assert.Equal(t, destinationName, storagePath)

	// Verify the file was copied
	fullPath, err := service.GetFullPath(storagePath)
	require.NoError(t, err)
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Equal(t, sourceContent, string(content))
}

func TestLocalStorageService_Delete(t *testing.T) {
	service, cleanup := setupTestStorage(t)
	defer cleanup()

	// First, save a file to delete
	sourceFile, err := os.CreateTemp("", "source-*.txt")
	require.NoError(t, err)
	os.Remove(sourceFile.Name()) // clean up source
	storagePath, err := service.Save(context.Background(), sourceFile.Name(), "file-to-delete.txt")
	require.NoError(t, err)

	fullPath, err := service.GetFullPath(storagePath)
	require.NoError(t, err)

	// Ensure it exists
	_, err = os.Stat(fullPath)
	require.NoError(t, err)

	// Now delete it
	err = service.Delete(context.Background(), storagePath)
	require.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err))
}

func TestLocalStorageService_GetFullPath(t *testing.T) {
	service, cleanup := setupTestStorage(t)
	defer cleanup()

	expectedPath := filepath.Join(service.basePath, "some-file.txt")
	fullPath, err := service.GetFullPath("some-file.txt")
	require.NoError(t, err)

	assert.Equal(t, expectedPath, fullPath)
}
