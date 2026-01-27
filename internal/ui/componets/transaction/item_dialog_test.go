package transaction

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestItemDialog_DefaultTaxLogic(t *testing.T) {
	app := test.NewApp()
	win := app.NewWindow("Test")

	t.Run("Default Tax 15% (Code 4)", func(t *testing.T) {
		dlg := NewItemDialog(win, func(item domain.TransactionItem) {}, 4)
		dlg.configureWidgets() 
		
		assert.Equal(t, "IVA 15%", dlg.taxSelect.Selected)
		assert.True(t, dlg.taxSelect.Disabled(), "Should be disabled when default is set")
	})

	t.Run("Default Tax 0% (Code 0)", func(t *testing.T) {
		dlg := NewItemDialog(win, func(item domain.TransactionItem) {}, 0)
		dlg.configureWidgets()
		
		assert.Equal(t, "IVA 0%", dlg.taxSelect.Selected)
		assert.True(t, dlg.taxSelect.Disabled())
	})

	t.Run("Default Tax Exempt (Code 7)", func(t *testing.T) {
		dlg := NewItemDialog(win, func(item domain.TransactionItem) {}, 7)
		dlg.configureWidgets()
		
		assert.Equal(t, "Exento (7)", dlg.taxSelect.Selected)
		assert.True(t, dlg.taxSelect.Disabled())
	})

	t.Run("Default Tax No Object (Code 6)", func(t *testing.T) {
		dlg := NewItemDialog(win, func(item domain.TransactionItem) {}, 6)
		dlg.configureWidgets()
		
		assert.Equal(t, "No Objeto (6)", dlg.taxSelect.Selected)
		assert.True(t, dlg.taxSelect.Disabled())
	})

	t.Run("No Default Tax (Code -1)", func(t *testing.T) {
		dlg := NewItemDialog(win, func(item domain.TransactionItem) {}, -1)
		dlg.configureWidgets()
		
		// Default UI behavior is usually 15% selected but ENABLED
		assert.Equal(t, "IVA 15%", dlg.taxSelect.Selected, "Should default to 15% if no preference")
		assert.False(t, dlg.taxSelect.Disabled(), "Should be enabled if no preference")
	})
}