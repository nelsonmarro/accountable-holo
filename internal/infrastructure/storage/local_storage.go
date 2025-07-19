package storage

import (
	"context"
	"fmt"
	"io"

	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// LocalStorageService implements the StorageService interface for saving files locally.
type LocalStorageService struct {
	app fyne.App
}

// NewLocalStorageService creates a new LocalStorageService using Fyne's storage API.
func NewLocalStorageService(a fyne.App) (*LocalStorageService, error) {
	// Ensure the attachments directory exists within the app's root storage.
	attachmentsURI, err := storage.Child(a.Storage().RootURI(), "attachments")
	if err != nil {
		return nil, fmt.Errorf("could not get attachments directory URI: %w", err)
	}
	exists, err := storage.Exists(attachmentsURI)
	if err != nil {
		return nil, fmt.Errorf("could not check for attachments directory: %w", err)
	}
	if !exists {
		if err := storage.CreateListable(attachmentsURI); err != nil {
			return nil, fmt.Errorf("could not create attachments directory: %w", err)
		}
	}
	return &LocalStorageService{app: a}, nil
}

// Save copies a file from a source URI to a permanent location.
func (s *LocalStorageService) Save(ctx context.Context, source fyne.URI, destinationName string) (string, error) {
	// Open the source file for reading
	sourceFile, err := storage.Reader(source)
	if err != nil {
		return "", fmt.Errorf("failed to open source file reader: %w", err)
	}
	defer sourceFile.Close()

	// Get the path to the app's attachment directory
	attachmentsURI, err := storage.Child(s.app.Storage().RootURI(), "attachments")
	if err != nil {
		return "", fmt.Errorf("could not get attachments directory URI for saving: %w", err)
	}

	// Create the destination file for writing
	destinationURI, err := storage.Child(attachmentsURI, destinationName)
	if err != nil {
		return "", fmt.Errorf("could not create destination URI: %w", err)
	}
	destinationFile, err := storage.Writer(destinationURI)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file writer: %w", err)
	}
	defer destinationFile.Close()

	// Copy the contents
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Return the string representation of the new URI to be stored in the DB
	return destinationURI.String(), nil
}

// GetFullPath returns the full, absolute path for a given URI string.
func (s *LocalStorageService) GetFullPath(storageURI string) (string, error) {
	uri, err := storage.ParseURI(storageURI)
	if err != nil {
		return "", fmt.Errorf("could not parse URI for GetFullPath: %w", err)
	}
	return uri.Path(), nil
}

// Delete removes a file from the storage directory.
func (s *LocalStorageService) Delete(ctx context.Context, storageURI string) error {
	uri, err := storage.ParseURI(storageURI)
	if err != nil {
		return fmt.Errorf("could not parse URI for deletion: %w", err)
	}
	return os.Remove(uri)
}
