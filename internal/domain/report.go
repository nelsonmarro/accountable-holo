package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type FinancialSummary struct {
	TotalIncome   decimal.Decimal
	TotalExpenses decimal.Decimal
	NetProfitLoss decimal.Decimal
}

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
