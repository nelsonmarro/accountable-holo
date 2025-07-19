package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2/storage"
	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type TransactionServiceImpl struct {
	repo    TransactionRepository
	storage StorageService
}

func NewTransactionService(repo TransactionRepository, storage StorageService) *TransactionServiceImpl {
	return &TransactionServiceImpl{repo: repo, storage: storage}
}

func (s *TransactionServiceImpl) GetTransactionByAccountPaginated(ctx context.Context, accountID, page, pageSize int, filter ...string) (*domain.PaginatedResult[domain.Transaction], error) {
	if accountID <= 0 {
		return nil, fmt.Errorf("ID de cuenta inválido: %d", accountID)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}
	if page > 100 {
		page = 100 // Limit to 100 pages
	}
	return s.repo.GetTransactionsByAccountPaginated(ctx, accountID, page, pageSize, filter...)
}

func (s *TransactionServiceImpl) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("ID de transacción inválido: %d", id)
	}
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la transacción: %w", err)
	}

	return tx, nil
}

func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transacción no puede ser nula")
	}
	txValidator := validator.New().For(tx)
	txValidator.Required("Amount", "Description", "TransactionDate", "AccountID", "CategoryID")
	txValidator.NumberMin(0, "Amount")
	txValidator.MaxDate(time.Now(), "TransactionDate")

	err := txValidator.ConsolidateErrors()
	if err != nil {
		return err
	}

	var sourceURIString string
	if tx.AttachmentPath != nil {
		sourceURIString = *tx.AttachmentPath
	}
	tx.AttachmentPath = nil // Clear it before initial creation

	err = s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("error al crear la transacción: %w", err)
	}

	if sourceURIString != "" {
		sourceURI, err := storage.ParseURI(sourceURIString)
		if err != nil {
			return fmt.Errorf("failed to parse source URI: %w", err)
		}

		destinationName := fmt.Sprintf("tx-%d-%s", tx.ID, sourceURI.Name())
		storageURI, err := s.storage.Save(ctx, sourceURI, destinationName)
		if err != nil {
			return fmt.Errorf("failed to save attachment: %w", err)
		}

		err = s.repo.UpdateAttachmentPath(ctx, tx.ID, storageURI)
		if err != nil {
			return fmt.Errorf("failed to update transaction with attachment path: %w", err)
		}
	}

	return nil
}

func (s *TransactionServiceImpl) VoidTransaction(ctx context.Context, transactionID int) error {
	if transactionID <= 0 {
		return fmt.Errorf("ID de transacción inválido: %d", transactionID)
	}

	return s.repo.VoidTransaction(ctx, transactionID)
}

func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transacción no puede ser nula")
	}

	txValidator := validator.New().For(tx)
	txValidator.Required("ID", "Amount", "Description", "TransactionDate", "AccountID", "CategoryID")
	txValidator.NumberMin(0, "Amount")
	txValidator.MaxDate(time.Now(), "TransactionDate")

	err := txValidator.ConsolidateErrors()
	if err != nil {
		return err
	}

	if tx.AttachmentPath != nil {
		sourceURIString := *tx.AttachmentPath
		sourceURI, err := storage.ParseURI(sourceURIString)
		if err != nil {
			return fmt.Errorf("failed to parse source URI for update: %w", err)
		}

		// Check if the path is a new file by checking if it's a local file path
		if _, err := os.Stat(sourceURI.Path()); err == nil {
			destinationName := fmt.Sprintf("tx-%d-%s", tx.ID, sourceURI.Name())
			storageURI, err := s.storage.Save(ctx, sourceURI, destinationName)
			if err != nil {
				return fmt.Errorf("failed to save new attachment: %w", err)
			}
			tx.AttachmentPath = &storageURI
		}
	}

	err = s.repo.UpdateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("error al actualizar la transacción: %w", err)
	}
	return nil
}