package componets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverableLabel is a custom widget that extends Label to show a tooltip on hover.
type HoverableLabel struct {
	widget.Label
	TooltipText string
	canvas      fyne.Canvas
	popup       *widget.PopUp
}

// NewHoverableLabel creates a new HoverableLabel widget.
func NewHoverableLabel(text string, canvas fyne.Canvas) *HoverableLabel {
	h := &HoverableLabel{
		canvas: canvas,
	}
	h.ExtendBaseWidget(h)
	h.SetText(text)
	return h
}

// MouseIn is called when the mouse enters the widget's area.
func (h *HoverableLabel) MouseIn(_ *desktop.MouseEvent) {
	if h.TooltipText != "" && h.popup == nil {
		content := container.NewPadded(widget.NewLabelWithStyle(h.TooltipText, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}))

		// Use a simple rectangle for background to make it look like a tooltip if needed,
		// but NewPopUp already has a background.

		h.popup = widget.NewPopUp(content, h.canvas)

		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		h.popup.ShowAtPosition(pos.Add(fyne.NewPos(0, h.Size().Height+5)))
	}
}

// MouseMoved is called when the mouse moves over the widget.
func (h *HoverableLabel) MouseMoved(_ *desktop.MouseEvent) {
}

// MouseOut is called when the mouse leaves the widget's area.
func (h *HoverableLabel) MouseOut() {
	if h.popup != nil {
		h.popup.Hide()
		h.popup = nil
	}
}

// SetTooltip sets the text to be displayed in the tooltip.
func (h *HoverableLabel) SetTooltip(text string) {
	h.TooltipText = text
}
