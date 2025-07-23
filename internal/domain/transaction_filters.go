package domain

import "time"

type TransactionFilters struct {
	StartDate    *time.Time
	EndDate      *time.Time
	CategoryID   *int
	CategoryType *CategoryType
	Description  *string
}
