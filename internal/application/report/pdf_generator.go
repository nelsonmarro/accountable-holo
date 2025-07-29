// Package report provides report generation.
package report

import (
	"context"
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
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
	// Configure the document with margins and page numbering
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	g.buildTitle(m, "Reporte de Transacciones")
	g.buildTransactionsTable(m, transactions)

	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
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
	// Define styles
	headerStyle := &props.Cell{BackgroundColor: &props.Color{Red: 220, Green: 230, Blue: 240}}
	zebraStyle := &props.Cell{BackgroundColor: &props.Color{Red: 245, Green: 245, Blue: 245}}
	voidedStyle := &props.Cell{BackgroundColor: &props.Color{Red: 255, Green: 220, Blue: 220}}  // Light red for voided rows
	voidingStyle := &props.Cell{BackgroundColor: &props.Color{Red: 220, Green: 230, Blue: 255}} // Light blue for voiding rows
	headerTextProps := props.Text{Style: fontstyle.Bold, Align: align.Center, Top: 2}
	cellTextProps := props.Text{Align: align.Center, Top: 2}

	headers := []string{"Fecha", "No.", "Categoría", "Tipo", "Monto", "Saldo"}

	// Build the table header
	m.AddRows(
		row.New(10).WithStyle(headerStyle).Add(
			text.NewCol(2, headers[0], headerTextProps),
			text.NewCol(2, headers[1], headerTextProps),
			text.NewCol(3, headers[2], headerTextProps),
			text.NewCol(1, headers[3], headerTextProps),
			text.NewCol(1, headers[4], headerTextProps),
			text.NewCol(3, headers[5], headerTextProps),
		),
	)

	// Add data rows
	for i, tx := range transactions {
		categoryName := ""
		categoryType := ""
		if tx.Category != nil {
			categoryName = tx.Category.Name
			categoryType = string(tx.Category.Type)
		}

		amountStyle := props.Text{Align: align.Center, Top: 2}
		if tx.Category.Type == domain.Income {
			amountStyle.Color = &props.GreenColor
		} else {
			amountStyle.Color = &props.RedColor
		}

		dataRow := row.New(10).Add(
			text.NewCol(2, tx.TransactionDate.Format("2006-01-02"), cellTextProps),
			text.NewCol(2, tx.TransactionNumber, cellTextProps),
			text.NewCol(3, categoryName, cellTextProps),
			text.NewCol(1, categoryType, cellTextProps),
			text.NewCol(1, fmt.Sprintf("%.2f", tx.Amount), amountStyle),
			text.NewCol(3, fmt.Sprintf("$%.2f", tx.RunningBalance), cellTextProps),
		)

		// Apply styles conditionally based on transaction state
		if tx.IsVoided {
			dataRow.WithStyle(voidedStyle)
		} else if tx.VoidsTransactionID != nil {
			dataRow.WithStyle(voidingStyle)
		} else if i%2 != 0 {
			dataRow.WithStyle(zebraStyle)
		}

		m.AddRows(dataRow)
	}
}
