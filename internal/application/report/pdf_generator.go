// Package report provides report generation.
package report

import (
	"context"

	"github.com/johnfercher/maroto/v2"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

// PDFReportGenerator generates reports in PDF format.
type PDFReportGenerator struct{}

// NewPDFReportGenerator creates a new instance of PDFReportGenerator.
func NewPDFReportGenerator() *PDFReportGenerator {
	return &PDFReportGenerator{}
}

// SelectedTransactionsReport generates a PDF report for selected transactions.
func (g *PDFReportGenerator) SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string) error {
	m := maroto.New()

	g.buildTitle(m, "Reporte de Transacciones")

	g.buildTransactionsTable(m, transactions)
}
