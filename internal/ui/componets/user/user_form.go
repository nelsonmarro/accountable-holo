package user

import (
	"fyne.io/fyne/v2/widget"
	"github.com/nelsonmarro/accountable-holo/internal/application/uivalidators"
)

func UserForm(
	username *widget.Entry,
	password *widget.Entry,
	role *widget.SelectEntry,
) []*widget.FormItem {
	addFormValidation(username, password, role)

	return []*widget.FormItem{
		{Text: "Username", Widget: username},
		{Text: "Password", Widget: password},
		{Text: "Role", Widget: role},
	}
}

func addFormValidation(
	username *widget.Entry,
	password *widget.Entry,
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

	roleValidator := uivalidators.NewValidator()
	roleValidator.Required()
	role.Validator = roleValidator.Validate
}
