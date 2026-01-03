package domain

import (
	"github.com/shopspring/decimal"
)

type CategoryAmount struct {
	CategoryName string
	Amount       decimal.Decimal
}

type BudgetStatus struct {
	CategoryName   string
	BudgetAmount   decimal.Decimal
	SpentAmount    decimal.Decimal
	PercentageUsed float64
	Remaining      decimal.Decimal
	IsOverBudget   bool
}

type FinancialSummary struct {
	TotalIncome        decimal.Decimal
	TotalExpenses      decimal.Decimal
	NetProfitLoss      decimal.Decimal
	IncomeByCategory   []CategoryAmount
	ExpensesByCategory []CategoryAmount
}
