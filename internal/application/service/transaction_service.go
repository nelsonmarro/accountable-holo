package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nelsonmarro/accountable-holo/internal/application/validator"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.com/shopspring/decimal"
)

type AccountService interface {
	GetAllAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccountByID(ctx context.Context, id int) (*domain.Account, error)
	CreateNewAccount(ctx context.Context, acc *domain.Account) error
	UpdateAccount(ctx context.Context, acc *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error
}

type TransactionServiceImpl struct {
	repo           TransactionRepository
	storage        StorageService
	accountService AccountService
}

func NewTransactionService(
	repo TransactionRepository,
	storage StorageService,
	accountService AccountService,
) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		repo:           repo,
		storage:        storage,
		accountService: accountService,
	}
}

func (s *TransactionServiceImpl) FindTransactionsByAccount(
	ctx context.Context,
	accountID int,
	page int,
	pageSize int,
	filters domain.TransactionFilters,
	searchString *string,
) (*domain.PaginatedResult[domain.Transaction], error) {
	result, err := s.repo.FindTransactionsByAccount(ctx, accountID, page, pageSize, filters, searchString)
	if err != nil {
		return nil, err
	}

	// Populate the absolute path for the UI
	for i := range result.Data {
		if result.Data[i].AttachmentPath != nil && *result.Data[i].AttachmentPath != "" {
			fullPath, err := s.storage.GetFullPath(*result.Data[i].AttachmentPath)
			if err == nil {
				result.Data[i].AbsoluteAttachPath = fullPath
			}
		}
	}

	return result, nil
}

func (s *TransactionServiceImpl) FindAllTransactionsByAccount(
	ctx context.Context,
	accountID int,
	filters domain.TransactionFilters,
) ([]domain.Transaction, error) {
	transactions, err := s.repo.FindAllTransactionsByAccount(ctx, accountID, filters, nil)
	if err != nil {
		return nil, err
	}

	// Populate the absolute path for the UI
	for i := range transactions {
		if transactions[i].AttachmentPath != nil && *transactions[i].AttachmentPath != "" {
			fullPath, err := s.storage.GetFullPath(*transactions[i].AttachmentPath)
			if err == nil {
				transactions[i].AbsoluteAttachPath = fullPath
			}
		}
	}

	return transactions, nil
}

func (s *TransactionServiceImpl) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("ID de transacción inválido: %d", id)
	}
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la transacción: %w", err)
	}

	// Populate the absolute path for the UI
	if tx.AttachmentPath != nil && *tx.AttachmentPath != "" {
		fullPath, err := s.storage.GetFullPath(*tx.AttachmentPath)
		if err == nil {
			tx.AbsoluteAttachPath = fullPath
		}
	}

	return tx, nil
}

func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, tx *domain.Transaction, currentUser domain.User) error {
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

	tx.CreatedByID = currentUser.ID
	tx.UpdatedByID = currentUser.ID

	var sourcePath string
	if tx.AttachmentPath != nil {
		sourcePath = *tx.AttachmentPath
	}
	tx.AttachmentPath = nil // Clear it before initial creation

	err = s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("error al crear la transacción: %w", err)
	}

	if sourcePath != "" {
		destinationName := fmt.Sprintf("tx-%d-%s", tx.ID, filepath.Base(sourcePath))
		storagePath, err := s.storage.Save(ctx, sourcePath, destinationName)
		if err != nil {
			return fmt.Errorf("failed to save attachment: %w", err)
		}

		err = s.repo.UpdateAttachmentPath(ctx, tx.ID, storagePath)
		if err != nil {
			return fmt.Errorf("failed to update transaction with attachment path: %w", err)
		}
	}

	return nil
}

func (s *TransactionServiceImpl) VoidTransaction(ctx context.Context, transactionID int, currentUser domain.User) error {
	if transactionID <= 0 {
		return fmt.Errorf("ID de transacción inválido: %d", transactionID)
	}

	return s.repo.VoidTransaction(ctx, transactionID, currentUser)
}

func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, tx *domain.Transaction, currentUser domain.User) error {
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

	tx.UpdatedByID = currentUser.ID

	if tx.AttachmentPath != nil {
		sourcePath := *tx.AttachmentPath
		if _, err := os.Stat(sourcePath); err == nil {
			destinationName := fmt.Sprintf("tx-%d-%s", tx.ID, filepath.Base(sourcePath))
			storagePath, err := s.storage.Save(ctx, sourcePath, destinationName)
			if err != nil {
				return fmt.Errorf("failed to save attachment: %w", err)
			}
			tx.AttachmentPath = &storagePath
		}
	}

	err = s.repo.UpdateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("error al actualizar la transacción: %w", err)
	}
	return nil
}

// ReconcileAccount reconciles the account transactions up to the specified end date.
func (s *TransactionServiceImpl) ReconcileAccount(
	ctx context.Context,
	accountID int,
	startDate time.Time,
	endDate time.Time,
	actualEndingBalance decimal.Decimal,
) (*domain.Reconciliation, error) {
	// Get the starting balance as of the start date
	startingBalance, err := s.repo.GetBalanceAsOf(ctx, accountID, startDate)
	if err != nil {
		return nil, fmt.Errorf("error al obtener el balance inicial: %w", err)
	}

	filters := domain.TransactionFilters{StartDate: &startDate, EndDate: &endDate}

	transactions, err := s.repo.FindAllTransactionsByAccount(ctx, accountID, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("error al obtener las transacciones de la cuenta: %w", err)
	}

	// Calculate the ending balance from the transactions
	calculatedEndingBalance := decimal.NewFromFloat(startingBalance)
	for _, tx := range transactions {
		amount := decimal.NewFromFloat(tx.Amount)
		if tx.Category.Type == domain.Income {
			calculatedEndingBalance = calculatedEndingBalance.Add(amount)
		} else {
			calculatedEndingBalance = calculatedEndingBalance.Sub(amount)
		}
	}

	discrepancy := actualEndingBalance.Sub(calculatedEndingBalance)

	// Assemble the reconciliation object
	reconciliation := &domain.Reconciliation{
		AccountID:               accountID,
		StartDate:               startDate,
		EndDate:                 endDate,
		StartingBalance:         decimal.NewFromFloat(startingBalance),
		CalculatedEndingBalance: calculatedEndingBalance,
		EndingBalance:           actualEndingBalance,
		Difference:              discrepancy,
		Transactions:            transactions,
	}

	return reconciliation, nil
}
