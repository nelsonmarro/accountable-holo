package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type CSVReportGenerator struct{}

func NewCSVReportGenerator() *CSVReportGenerator {
	return &CSVReportGenerator{}
}

func (g *CSVReportGenerator) SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Número de Transacción", "Descripción", "Monto", "Fecha de Transacción", "Nombre de Categoría", "Tipo de Categoría", "Saldo Acumulado"}
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
			tx.TransactionNumber,
			tx.Description,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.TransactionDate.Format("2006-01-02"),
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
