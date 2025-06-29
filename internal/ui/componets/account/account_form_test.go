package account

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddFormValidation(t *testing.T) {
	t.Run("should have validators set on every entry", func(t *testing.T) {
		// Arrange
		nameEntry := widget.NewEntry()
		amountEntry := widget.NewEntry()
		tipoSelect := widget.NewSelectEntry([]string{"Ahorros", "Corriente"})
		numberEntry := widget.NewEntry()

		// Act
		addFormValidation(nameEntry, amountEntry, tipoSelect, numberEntry)

		// Assert
		assert.NotEmpty(t, nameEntry.Validator, "Name entry should have a validator")
		assert.NotEmpty(t, amountEntry.Validator, "Amount entry should have a validator")
		assert.NotEmpty(t, tipoSelect.Validator, "Tipo select should have a validator")
		assert.NotEmpty(t, numberEntry.Validator, "Number entry should have a validator")
	})

	t.Run("should rise no validator error if input is valid", func(t *testing.T) {
		// Arrange
		nameEntry := widget.NewEntry()
		amountEntry := widget.NewEntry()
		tipoSelect := widget.NewSelectEntry([]string{"Ahorros", "Corriente"})
		numberEntry := widget.NewEntry()

		// Act
		addFormValidation(nameEntry, amountEntry, tipoSelect, numberEntry)

		// Assert
		test.Type(nameEntry, "Test Bank Account")
		test.Type(amountEntry, "150.75")
		test.Type(tipoSelect, "Ahorros")
		test.Type(numberEntry, "2222")

		require.NoError(t, nameEntry.Validate(), "Name entry should be valid")
		require.NoError(t, amountEntry.Validate(), "Amount entry should be valid")
		require.NoError(t, tipoSelect.Validate(), "Tipo select should be valid")
		require.NoError(t, numberEntry.Validate(), "Number entry should be valid")
	})
}
