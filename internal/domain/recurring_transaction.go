package domain

import "time"

// RecurrenceInterval represents the interval at which a recurring transaction occurs.
type RecurrenceInterval string

const (
	IntervalMonthly RecurrenceInterval = "MONTHLY"
	IntervalWeekly  RecurrenceInterval = "WEEKLY"
)

type RecurringTransaction struct {
	BaseEntity
	Description string  `db:"description"`
	Amount      float64 `db:"amount"`
	AccountID   int     `db:"account_id"`
	CategoryID  int     `db:"category_id"`

	// Recurrence rules
	Interval    RecurrenceInterval `db:"interval"`
	StartDate   time.Time          `db:"start_date"`
	NextRunDate time.Time          `db:"next_run_date"`
	IsActive    bool               `db:"is_active"`

	Account  *Account  `db:"-"`
	Category *Category `db:"-"`
}
