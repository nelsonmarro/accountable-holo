package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestService is a helper function to create a new service in a temporary directory.
func setupTestService(t *testing.T) (*LocalStorageService, string) {
	// Use a temporary directory for the "user home" to isolate the test
	tempHome := t.TempDir()
	attachmentsDir := "test_attachments"

	// We need to temporarily set the user home dir for the test
	// This is a bit of a hack, but it's the cleanest way to test this.
	originalHome, err := os.UserHomeDir()
	require.NoError(t, err)
	t.Setenv("HOME", tempHome)        // For Linux/macOS
	t.Setenv("USERPROFILE", tempHome) // For Windows

	// Restore the original home dir after the test
	t.Cleanup(func() {
		t.Setenv("HOME", originalHome)
		t.Setenv("USERPROFILE", originalHome)
	})

	service, err := NewLocalStorageService(attachmentsDir)
	require.NoError(t, err, "NewLocalStorageService should not return an error")
	require.NotNil(t, service, "Service should not be nil")

	expectedPath := filepath.Join(tempHome, "verith", attachmentsDir)
	return service, expectedPath
}

func TestNewLocalStorageService(t *testing.T) {
	service, expectedPath := setupTestService(t)

	// Assert that the base path is correct
	assert.Equal(t, expectedPath, service.basePath, "BasePath should be set correctly")

	// Assert that the directory was actually created
	_, err := os.Stat(expectedPath)
	assert.NoError(t, err, "Storage directory should exist")
}

func TestLocalStorageService_Save(t *testing.T) {
	service, storagePath := setupTestService(t)
	ctx := context.Background()

	// 1. Create a temporary source file
	sourceFile, err := os.CreateTemp("", "source_*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(sourceFile.Name()) }()

	testContent := "hello world"
	_, err = sourceFile.WriteString(testContent)
	require.NoError(t, err)
	_ = sourceFile.Close()

	// 2. Call the Save method
	destinationName := "my_saved_file.txt"
	savedPath, err := service.Save(ctx, sourceFile.Name(), destinationName)
	require.NoError(t, err)

	// 3. Assertions
	assert.Equal(t, destinationName, savedPath, "Save should return the destination name")

	fullDestinationPath := filepath.Join(storagePath, destinationName)
	content, err := os.ReadFile(fullDestinationPath)
	require.NoError(t, err, "Saved file should be readable")
	assert.Equal(t, testContent, string(content), "Content of saved file should match source")
}

func TestLocalStorageService_GetFullPath(t *testing.T) {
	service, storagePath := setupTestService(t)
	relativePath := "my_file.txt"
	expectedPath := filepath.Join(storagePath, relativePath)

	fullPath, err := service.GetFullPath(relativePath)
	require.NoError(t, err)
	assert.Equal(t, expectedPath, fullPath, "GetFullPath should construct the correct absolute path")
}

func TestLocalStorageService_Delete(t *testing.T) {
	service, storagePath := setupTestService(t)
	ctx := context.Background()

	// 1. First, save a file so we can delete it
	fileName := "file_to_delete.txt"
	filePath := filepath.Join(storagePath, fileName)
	err := os.WriteFile(filePath, []byte("delete me"), 0o666)
	require.NoError(t, err)

	// 2. Call the Delete method
	err = service.Delete(ctx, fileName)
	require.NoError(t, err, "Delete should not return an error for an existing file")

	// 3. Assert the file is gone
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err), "File should no longer exist after deletion")

	// 4. Test deleting a non-existent file
	err = service.Delete(ctx, "non_existent_file.txt")
	assert.NoError(t, err, "Deleting a non-existent file should not return an error")
}
