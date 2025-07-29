// Package report provides report generation.
package report

import (
	"context"
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
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

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF", err)
	}

	return document.Save(outputPath)
}

func (g *PDFReportGenerator) buildTitle(m core.Maroto, title string) {
	m.AddRow(
		15,
		col.New(12).Add(
			text.New(title, props.Text{
				Top:   5,
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)
}

func (g *PDFReportGenerator) buildTransactionsTable(m core.Maroto, transactions []domain.Transaction) {
	headers := []string{"Fecha", "No.", "Descripción", "Categoría", "Tipo", "Monto", "Saldo"}

	// Build the table header
	m.AddRow(10,
		col.New(2).Add(text.New(headers[0], props.Text{Style: fontstyle.Bold, Align: align.Center})),
		col.New(1).Add(text.New(headers[1], props.Text{Style: fontstyle.Bold, Align: align.Center})),
		col.New(3).Add(text.New(headers[2], props.Text{Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New(headers[3], props.Text{Style: fontstyle.Bold, Align: align.Center})),
		col.New(1).Add(text.New(headers[4], props.Text{Style: fontstyle.Bold, Align: align.Center})),
		col.New(2).Add(text.New(headers[5], props.Text{Style: fontstyle.Bold, Align: align.Right})),
		col.New(1).Add(text.New(headers[6], props.Text{Style: fontstyle.Bold, Align: align.Right})),
	)
}
