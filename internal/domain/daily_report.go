package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// DailyReport holds the aggregated data for a single day's financial activity.
type DailyReport struct {
	AccountID       int
	ReportDate      time.Time
	CurrentBalance  decimal.Decimal
	DailyIncome     decimal.Decimal
	DailyExpenses   decimal.Decimal
	DailyProfitLoss decimal.Decimal
	Transactions    []Transaction
}
