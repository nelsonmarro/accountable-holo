package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorageService implements the StorageService interface for saving files locally.
type LocalStorageService struct {
	storagePath string
}

// NewLocalStorageService creates a new LocalStorageService.
// It also ensures the base storage directory exists.
func NewLocalStorageService(path string) (*LocalStorageService, error) {
	if path == "" {
		return nil, fmt.Errorf("storage path cannot be empty")
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageService{storagePath: path}, nil
}

// Save copies a file from sourcePath to a permanent location within the storagePath.
func (s *LocalStorageService) Save(ctx context.Context, sourcePath, destinationName string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationPath := filepath.Join(s.storagePath, destinationName)
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Return the path relative to the execution directory
	return destinationPath, nil
}

// GetFullPath converts a path stored in the database to a full, absolute path.
func (s *LocalStorageService) GetFullPath(storagePath string) (string, error) {
	return filepath.Abs(storagePath)
}

// Delete removes a file from the storage directory.
func (s *LocalStorageService) Delete(ctx context.Context, storagePath string) error {
	fullPath, err := s.GetFullPath(storagePath)
	if err != nil {
		return fmt.Errorf("could not get full path for deletion: %w", err)
	}
	return os.Remove(fullPath)
}