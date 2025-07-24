package report

import (
	"context"
	"fmt"

	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type PDFReportGenerator struct{}

func NewPDFReportGenerator() *PDFReportGenerator {
	return &PDFReportGenerator{}
}

func (g *PDFReportGenerator) Generate(ctx context.Context, transactions []domain.Transaction, outputPath string) error {
	// TODO: Implement PDF generation logic here
	return fmt.Errorf("PDF generation not yet implemented")
}
