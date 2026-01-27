package componets

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// LatinDateEntry is a custom widget that forces DD/MM/YYYY format
// regardless of the system locale. It combines an Entry with a Calendar button.
type LatinDateEntry struct {
	widget.BaseWidget
	
	Entry *widget.Entry
	Date  *time.Time // The selected date (nil if empty/invalid)
	
	parentWindow fyne.Window
	onChanged    func(time.Time)
}

func NewLatinDateEntry(parent fyne.Window) *LatinDateEntry {
	e := &LatinDateEntry{
		parentWindow: parent,
		Entry:        widget.NewEntry(),
	}
	e.Entry.SetPlaceHolder("DD/MM/YYYY")
	
	// Validation Logic inline for clarity
	e.Entry.Validator = func(s string) error {
		if s == "" {
			return nil // Allow empty? Or required? Let consumer decide via external validator.
		}
		_, err := time.Parse(AppDateFormat, s)
		return err
	}

	e.Entry.OnChanged = func(s string) {
		t, err := time.Parse(AppDateFormat, s)
		if err == nil {
			e.Date = &t
			if e.onChanged != nil {
				e.onChanged(t)
			}
		} else {
			e.Date = nil
		}
	}

	e.ExtendBaseWidget(e)
	return e
}

// SetDate sets the date and updates the text field.
func (e *LatinDateEntry) SetDate(t time.Time) {
	e.Date = &t
	e.Entry.SetText(t.Format(AppDateFormat))
}

// SetText manually sets the text (and updates Date if valid)
func (e *LatinDateEntry) SetText(s string) {
	e.Entry.SetText(s)
	// OnChanged triggers automatically? Fyne Entry usually triggers OnChanged when SetText is called?
	// Actually, usually NO. We must trigger logic manually if needed.
	// But let's check validation.
	t, err := time.Parse(AppDateFormat, s)
	if err == nil {
		e.Date = &t
	} else {
		e.Date = nil
	}
}

func (e *LatinDateEntry) CreateRenderer() fyne.WidgetRenderer {
	icon := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		e.showCalendar()
	})
	
	// Use a Border layout: Button on Right, Entry fills the rest
	c := container.NewBorder(nil, nil, nil, icon, e.Entry)
	
	return widget.NewSimpleRenderer(c)
}

func (e *LatinDateEntry) showCalendar() {
	var d dialog.Dialog

	startDate := time.Now()
	if e.Date != nil {
		startDate = *e.Date
	}

	// Create a calendar widget
	cal := widget.NewCalendar(startDate, func(t time.Time) {
		e.SetDate(t)
		if d != nil {
			d.Hide()
		}
	})

	// Wrap in a custom dialog
	d = dialog.NewCustom("Seleccionar Fecha", "Cerrar", container.NewPadded(cal), e.parentWindow)
	d.Show()
}
