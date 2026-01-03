package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDatesForRange(t *testing.T) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	tests := []struct {
		name          string
		rangeOption   string
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:        "Este Mes",
			rangeOption: "Este Mes",
			expectedStart: time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.Local),
			expectedEnd:   time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1),
		},
		{
			name:        "Mes Pasado",
			rangeOption: "Mes Pasado",
			expectedStart: time.Date(currentYear, currentMonth-1, 1, 0, 0, 0, 0, time.Local),
			expectedEnd:   time.Date(currentYear, currentMonth-1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1),
		},
		{
			name:        "Este Año",
			rangeOption: "Este Año",
			expectedStart: time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.Local),
			expectedEnd:   time.Date(currentYear, 12, 31, 0, 0, 0, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getDatesForRange(tt.rangeOption)
			
			// We compare only Year, Month, Day to avoid issues with exact time due to "now"
			assert.Equal(t, tt.expectedStart.Format("2006-01-02"), start.Format("2006-01-02"))
			assert.Equal(t, tt.expectedEnd.Format("2006-01-02"), end.Format("2006-01-02"))
		})
	}
}

func TestGetDatesForRange_EsteTrimestre(t *testing.T) {
	// Special case for Quarter because it depends on the current month dynamically
	now := time.Now()
	currentMonth := now.Month()
	quarter := (int(currentMonth) - 1) / 3
	startMonth := time.Month(quarter*3 + 1)
	
	expectedStart := time.Date(now.Year(), startMonth, 1, 0, 0, 0, 0, time.Local)
	expectedEnd := expectedStart.AddDate(0, 3, -1)

	start, end := getDatesForRange("Este Trimestre")

	assert.Equal(t, expectedStart.Format("2006-01-02"), start.Format("2006-01-02"))
	assert.Equal(t, expectedEnd.Format("2006-01-02"), end.Format("2006-01-02"))
}
