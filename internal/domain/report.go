package domain

import (
	"github.com/shopspring/decimal"
)

type FinancialSummary struct {
	TotalIncome   decimal.Decimal
	TotalExpenses decimal.Decimal
	NetProfitLoss decimal.Decimal
}
