// Package report provides report generation.
package report

import (
	"context"
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/breakline"
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
	headerStyle := props.Cell{
		BackgroundColor: &props.Color{
			Red:   220,
			Green: 230,
			Blue:  240,
		},
	}

	headerTextProps := props.Text{
		Style: fontstyle.Bold,
		Align: align.Center,
		Top:   2,
	}

	headers := []string{"Fecha", "No.", "Descripción", "Categoría", "Tipo", "Monto", "Saldo"}

	// Build the table header
	m.AddRows(
		row.New(10).WithStyle(&headerStyle).Add(
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
	for _, tx := range transactions {
		categoryName := ""
		categoryType := ""
		if tx.Category != nil {
			categoryName = tx.Category.Name
			categoryType = string(tx.Category.Type)
		}

		amountStyle := props.Text{
			Align: align.Center,
			Top:   2,
		}
		if tx.Category.Type == domain.Income {
			amountStyle.Color = &props.GreenColor
		} else {
			amountStyle.Color = &props.RedColor
		}

		descriptionStyle := props.Text{
			Align:             align.Justify,
			Top:               2,
			BreakLineStrategy: breakline.EmptySpaceStrategy,
		}

		// Add a row for the transaction data
		m.AddRow(10,
			col.New(2).Add(text.New(tx.TransactionDate.Format("2006-01-02"), props.Text{Align: align.Center, Top: 2})),
			col.New(1).Add(text.New(tx.TransactionNumber, props.Text{Align: align.Center, Top: 2})),
			col.New(3).Add(text.New(tx.Description, descriptionStyle)),
			col.New(2).Add(text.New(categoryName, props.Text{Align: align.Center, Top: 2})),
			col.New(1).Add(text.New(categoryType, props.Text{Align: align.Center, Top: 2})),
			col.New(2).Add(text.New(fmt.Sprintf("%.2f", tx.Amount), amountStyle)),
			col.New(1).Add(text.New(fmt.Sprintf("%.2f", tx.RunningBalance), props.Text{Align: align.Center, Top: 2})),
		)

		// Add a separator line
		m.AddRow(5,
			col.New(12).Add(line.New(props.Line{Thickness: 0.1})),
		)
	}
}
