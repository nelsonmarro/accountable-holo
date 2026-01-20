package user

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/verith/internal/application/uivalidators"
)

func UserForm(
	username *widget.Entry,
	password *widget.Entry,
	firstName *widget.Entry,
	lastName *widget.Entry,
	role *widget.SelectEntry,
) []*widget.FormItem {
	addFormValidation(username, password, firstName, lastName, role)

	return []*widget.FormItem{
		{Text: "Username", Widget: username},
		{Text: "Password", Widget: password},
		{Text: "First Name", Widget: firstName},
		{Text: "Last Name", Widget: lastName},
		{Text: "Role", Widget: role},
	}
}

func addFormValidation(
	username *widget.Entry,
	password *widget.Entry,
	firstName *widget.Entry,
	lastName *widget.Entry,
	role *widget.SelectEntry,
) {
	usernameValidator := uivalidators.NewValidator()
	usernameValidator.Required()
	usernameValidator.MinLength(3)
	username.Validator = usernameValidator.Validate

	passwordValidator := uivalidators.NewValidator()
	passwordValidator.Required()
	passwordValidator.MinLength(8)
	password.Validator = passwordValidator.Validate

	firstNameValidator := uivalidators.NewValidator()
	firstNameValidator.Required()
	firstName.Validator = firstNameValidator.Validate

	lastNameValidator := uivalidators.NewValidator()
	lastNameValidator.Required()
	lastName.Validator = lastNameValidator.Validate

	roleValidator := uivalidators.NewValidator()
	roleValidator.Required()
	role.Validator = roleValidator.Validate
}
