package account

import (
	"testing"

	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestAddFormValidation(t *testing.T) {
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
}
