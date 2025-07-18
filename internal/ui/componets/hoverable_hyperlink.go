package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type HoverableHyperLink struct {
	widget.Hyperlink
	TooltipText string
	canvas      fyne.Canvas
	popup       *widget.PopUp
}
