package componets

import (
	"net/url"
	"testing"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestNewHoverableHyperlink(t *testing.T) {
	// Setup a test Fyne app
	a := test.NewApp()
	w := a.NewWindow("Test")
	defer w.Close()

	testURL, _ := url.Parse("https://example.com")
	hyperlink := NewHoverableHyperlink("Click Me", testURL, w.Canvas())

	assert.NotNil(t, hyperlink, "Hyperlink should not be nil")
	assert.Equal(t, "Click Me", hyperlink.Text, "Hyperlink text should be set correctly")
	assert.Equal(t, testURL, hyperlink.URL, "Hyperlink URL should be set correctly")
}

func TestHoverableHyperlink_Tooltip(t *testing.T) {
	a := test.NewApp()
	w := a.NewWindow("Test")
	defer w.Close()

	testURL, _ := url.Parse("https://example.com")
	hyperlink := NewHoverableHyperlink("Click Me", testURL, w.Canvas())
	hyperlink.Resize(hyperlink.MinSize())

	// Set the tooltip text
	hyperlink.SetTooltip("This is a tooltip")
	assert.Equal(t, "This is a tooltip", hyperlink.TooltipText)

	// Simulate mouse entering the widget
	hyperlink.MouseIn(&desktop.MouseEvent{})
	assert.NotNil(t, hyperlink.popup, "Popup should be visible on MouseIn")
	assert.True(t, hyperlink.popup.Visible(), "Popup should be visible")

	// Simulate mouse leaving the widget
	hyperlink.MouseOut()
	assert.Nil(t, hyperlink.popup, "Popup should be nil after MouseOut")
}
