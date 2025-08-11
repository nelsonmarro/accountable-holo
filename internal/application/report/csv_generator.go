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

func (g *CSVReportGenerator) SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string, currentUser *domain.User) error {
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

	// Add footer
	writer.Write([]string{}) // Spacer
	writer.Write([]string{"Reporte Generado Por:", fmt.Sprintf("%s %s", currentUser.FirstName, currentUser.LastName)})

	return nil
}

func (g *CSVReportGenerator) DailyReport(ctx context.Context, report *domain.DailyReport, outputPath string, currentUser *domain.User) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Summary Header
	if err := writer.Write([]string{"Reporte Financiero Diario"}); err != nil {
		return err
	}
	// Summary Data
	summaryData := [][]string{
		{"Fecha del Reporte", report.ReportDate.Format("2006-01-02 15:04:05")},
		{"Saldo Actual", report.CurrentBalance.StringFixed(2)},
		{"Ingresos del Día", report.DailyIncome.StringFixed(2)},
		{"Egresos del Día", report.DailyExpenses.StringFixed(2)},
		{"Ganancia/Pérdida Neta", report.DailyProfitLoss.StringFixed(2)},
	}
	if err := writer.WriteAll(summaryData); err != nil {
		return err
	}

	// Spacer
	if err := writer.Write([]string{}); err != nil {
		return err
	}

	// Transaction Header
	txHeader := []string{"Número de Transacción", "Descripción", "Monto", "Fecha de Transacción", "Nombre de Categoría", "Tipo de Categoría"}
	if err := writer.Write(txHeader); err != nil {
		return err
	}

	// Transaction Data
	for _, tx := range report.Transactions {
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
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Add footer
	writer.Write([]string{}) // Spacer
	writer.Write([]string{"Reporte Generado Por:", fmt.Sprintf("%s %s", currentUser.FirstName, currentUser.LastName)})

	return nil
}
