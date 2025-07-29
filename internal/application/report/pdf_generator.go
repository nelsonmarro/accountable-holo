// Package report provides report generation.
package report

import (
	"context"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
)

type PDFReportGenerator struct{}

func NewPDFReportGenerator() *PDFReportGenerator {
	return &PDFReportGenerator{}
}

func (g *PDFReportGenerator) SelectedTransactionsReport(ctx context.Context, transactions []domain.Transaction, outputPath string) error {
	m := maroto.New(config.New())
}
