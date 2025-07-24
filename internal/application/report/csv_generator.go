package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type ReportGenerator interface {
	Generate(ctx context.Context, transactions []domain.Transaction, outputPath string) error
}

type CSVReportGenerator struct{}

func NewCSVReportGenerator() *CSVReportGenerator {
	return &CSVReportGenerator{}
}

func (g *CSVReportGenerator) Generate(ctx context.Context, transactions []domain.Transaction, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"ID", "Transaction Number", "Description", "Amount", "Transaction Date", "Account ID", "Category ID", "Category Name", "Category Type", "Running Balance"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, tx := range transactions {
		categoryName := ""
		categoryType := ""
		if tx.Category != nil {
			categoryName = tx.Category.Name
			categoryType = string(tx.Category.Type)
		}

		record := []string{
			fmt.Sprintf("%d", tx.ID),
			tx.TransactionNumber,
			tx.Description,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.TransactionDate.Format("2006-01-02"),
			fmt.Sprintf("%d", tx.AccountID),
			fmt.Sprintf("%d", tx.CategoryID),
			categoryName,
			categoryType,
			fmt.Sprintf("%.2f", tx.RunningBalance),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}
