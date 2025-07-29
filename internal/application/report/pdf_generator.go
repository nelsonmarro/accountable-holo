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
	"github.com/johnfercher/maroto/v2/pkg/consts/breakline"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/nelsonmarro/accountable-holo/internal/domain"
	"github.comcom/johnfercher/maroto/v2/pkg/props"
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
	zebraStyle := &props.Cell{BackgroundColor: &props.Color{Red: 235, Green: 235, Blue: 235}}
	headerTextProps := props.Text{Style: fontstyle.Bold, Align: align.Center, Top: 2}
	cellTextProps := props.Text{Align: align.Center, Top: 2}
	descriptionStyle := props.Text{Align: align.Left, Top: 2, BreakLineStrategy: breakline.EmptySpaceStrategy}
	headers := []string{"Fecha", "No.", "Descripción", "Categoría", "Tipo", "Monto", "Saldo"}

	// Build the table header
	m.AddRows(
		row.New(10).WithStyle(headerStyle).Add(
			text.NewCol(2, headers[0], headerTextProps),
			text.NewCol(1, headers[1], headerTextProps),
			text.NewCol(3, headers[2], headerTextProps),
			text.NewCol(2, headers[3], headerTextProps),
			text.NewCol(1, headers[4], headerTextProps),
			text.NewCol(2, headers[5], headerTextProps),
			text.NewCol(1, headers[6], headerTextProps),
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

		// Add a row with automatic height calculation
		dataRow := m.AddAutoRow(
			col.New(2).Add(text.New(tx.TransactionDate.Format("2006-01-02"), cellTextProps)),
			col.New(1).Add(text.New(tx.TransactionNumber, cellTextProps)),
			col.New(3).Add(text.New(tx.Description, descriptionStyle)),
			col.New(2).Add(text.New(categoryName, cellTextProps)),
			col.New(1).Add(text.New(categoryType, cellTextProps)),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", tx.Amount), amountStyle)),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", tx.RunningBalance), cellTextProps)),
		)

		// Apply zebra striping to even rows
		if i%2 == 0 {
			dataRow.WithStyle(zebraStyle)
		}
	}
}

