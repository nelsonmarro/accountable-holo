package persistence

import "github.com/jackc/pgx/v5/pgxpool"

type TransactionRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{db: db}
}

