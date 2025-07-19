package componets

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverableHyperlink is a custom widget that extends Hyperlink to show a tooltip on hover.
type HoverableHyperlink struct {
	widget.Hyperlink
	TooltipText string
	canvas      fyne.Canvas
	popup       *widget.PopUp
}

// NewHoverableHyperlink creates a new HoverableHyperlink widget.
func NewHoverableHyperlink(text string, url *url.URL, canvas fyne.Canvas) *HoverableHyperlink {
	h := &HoverableHyperlink{
		canvas: canvas,
	}
	h.ExtendBaseWidget(h) // This is crucial for custom widgets
	h.SetText(text)
	h.SetURL(url)
	return h
}

// MouseIn is called when the mouse enters the widget's area.
func (h *HoverableHyperlink) MouseIn(_ *desktop.MouseEvent) {
	if h.TooltipText != "" && h.popup == nil {
		label := widget.NewLabel(h.TooltipText)
		h.popup = widget.NewPopUp(label, h.canvas)
		// Position the popup below the hyperlink
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		h.popup.ShowAtPosition(pos.Add(fyne.NewPos(0, h.Size().Height)))
	}
}

// MouseMoved is called when the mouse moves over the widget.
func (h *HoverableHyperlink) MouseMoved(_ *desktop.MouseEvent) {
	// We must implement this to satisfy the desktop.Hoverable interface
}

// MouseOut is called when the mouse leaves the widget's area.
func (h *HoverableHyperlink) MouseOut() {
	if h.popup != nil {
		h.popup.Hide()
		h.popup = nil
	}
}

// SetTooltip sets the text to be displayed in the tooltip.
func (h *HoverableHyperlink) SetTooltip(text string) {
	h.TooltipText = text
}
