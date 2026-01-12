// Package storage provides a local file storage service for saving, retrieving, and deleting files.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorageService struct {
	basePath string
}

func NewLocalStorageService(attachmentsDir string) (*LocalStorageService, error) {
	userConfigDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	storagePath := filepath.Join(userConfigDir, "accountable-holo", attachmentsDir)
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageService{basePath: storagePath}, nil
}

func (s *LocalStorageService) Save(ctx context.Context, sourcePath string, destinationName string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }()

	destinationPath := filepath.Join(s.basePath, destinationName)
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destinationFile.Close() }()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return destinationName, nil
}

func (s *LocalStorageService) GetFullPath(storagePath string) (string, error) {
	return filepath.Join(s.basePath, storagePath), nil
}

func (s *LocalStorageService) Delete(ctx context.Context, storagePath string) error {
	fullPath := filepath.Join(s.basePath, storagePath)
	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // If the file doesn't exist, it's already "deleted".
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
