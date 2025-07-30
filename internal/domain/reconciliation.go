package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Reconciliation struct {
	AccountID               int
	StartDate               time.Time
	EndDate                 time.Time
	StartingBalance         decimal.Decimal
	EndingBalance           decimal.Decimal
	CalculatedEndingBalance decimal.Decimal
	Difference              decimal.Decimal
	Transactions            []Transaction
}

