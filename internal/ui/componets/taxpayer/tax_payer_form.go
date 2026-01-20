package taxpayer

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/uivalidators"
	"github.com/nelsonmarro/verith/internal/domain"
)

type TaxPayerForm struct {
	NameEntry    *widget.Entry
	IdEntry      *widget.Entry
	EmailEntry   *widget.Entry
	AddrEntry    *widget.Entry
	PhoneEntry   *widget.Entry
	FormWidget   *widget.Form
}

func NewTaxPayerForm() *TaxPayerForm {
	f := &TaxPayerForm{
		NameEntry:  widget.NewEntry(),
		IdEntry:    widget.NewEntry(),
		EmailEntry: widget.NewEntry(),
		AddrEntry:  widget.NewEntry(),
		PhoneEntry: widget.NewEntry(),
	}

	f.NameEntry.SetPlaceHolder("Razón Social / Nombre")
	f.IdEntry.SetPlaceHolder("RUC o Cédula (10 o 13 dígitos)")
	f.EmailEntry.SetPlaceHolder("correo@ejemplo.com")
	f.AddrEntry.SetPlaceHolder("Dirección completa")
	f.PhoneEntry.SetPlaceHolder("Teléfono (Opcional)")

	// --- Validación ---
	idVal := uivalidators.NewValidator()
	idVal.Required()
	idVal.IsInt()
	idVal.MinLength(10)
	f.IdEntry.Validator = idVal.Validate

	nameVal := uivalidators.NewValidator()
	nameVal.Required()
	nameVal.MinLength(3)
	f.NameEntry.Validator = nameVal.Validate

	emailVal := uivalidators.NewValidator()
	emailVal.Required()
	// uivalidators doesn't seem to have IsEmail yet based on previous context, 
	// assuming Basic check or just Required for now.
	f.EmailEntry.Validator = emailVal.Validate

	f.FormWidget = widget.NewForm(
		widget.NewFormItem("Identificación", f.IdEntry),
		widget.NewFormItem("Nombre", f.NameEntry),
		widget.NewFormItem("Email", f.EmailEntry),
		widget.NewFormItem("Dirección", f.AddrEntry),
		widget.NewFormItem("Teléfono", f.PhoneEntry),
	)

	return f
}

func (f *TaxPayerForm) LoadData(tp *domain.TaxPayer) {
	f.NameEntry.SetText(tp.Name)
	f.IdEntry.SetText(tp.Identification)
	f.EmailEntry.SetText(tp.Email)
	f.AddrEntry.SetText(tp.Address)
	f.PhoneEntry.SetText(tp.Phone)
	f.IdEntry.Disable() // ID shouldn't change generally
}

func (f *TaxPayerForm) GetTaxPayer() *domain.TaxPayer {
	tp := &domain.TaxPayer{
		Name:           f.NameEntry.Text,
		Identification: f.IdEntry.Text,
		Email:          f.EmailEntry.Text,
		Address:        f.AddrEntry.Text,
		Phone:          f.PhoneEntry.Text,
		IdentificationType: "04", // RUC default
	}

	if len(tp.Identification) == 10 {
		tp.IdentificationType = "05" // Cédula
	}
	return tp
}
