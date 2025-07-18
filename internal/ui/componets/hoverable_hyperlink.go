package componets

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type HoverableHyperLink struct {
	widget.Hyperlink
	TooltipText string
	canvas      fyne.Canvas
	popup       *widget.PopUp
}

func NewHoverableHyperLink(text string, url *url.URL, canvas fyne.Canvas) *HoverableHyperLink {
	h := &HoverableHyperLink{
		canvas: canvas,
	}
	h.ExtendBaseWidget(h)
	h.SetText(text)
	h.SetURL(url)
	return h
}

// MouseIn is called when the mouse enters the widget area.
func (h *HoverableHyperLink) MouseIn(_ *desktop.MouseEvent) {
	if h.TooltipText != "" && h.popup == nil {
		label := widget.NewLabel(h.TooltipText)
		h.popup = widget.NewPopUp(label, h.canvas)

		// Position the popup below the hyperlink
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(h)
		h.popup.ShowAtPosition(pos.Add(fyne.NewPos(0, h.Size().Height)))
	}
}

func (h *HoverableHyperLink) MouseMoved(_ *desktop.MouseEvent) {
	// We need to implement this to satisfy the desktop.Hoverable interface,
}
