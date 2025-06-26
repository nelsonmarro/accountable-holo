package service

import (
	"context"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type AccountRepository interface {
	GetAllAccounts(ctx context.Context) ([]domain.Account, error)
	GetAccountByID(ctx context.Context, id int) (*domain.Account, error)
	AccountExists(ctx context.Context, id int) (bool, error)
	CreateAccount(ctx context.Context, acc *domain.Account) error
	UpdateAccount(ctx context.Context, acc *domain.Account) error
	DeleteAccount(ctx context.Context, id int) error
}
