package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeAwareResource is a struct that holds two resources, one for light theme and one for dark theme.
type ThemeAwareResource struct {
	LightResource fyne.Resource
	DarkResource  fyne.Resource
}

// NewThemeAwareResource is a constructor for ThemeAwareResource.
func NewThemeAwareResource(lightResource, darkResource fyne.Resource) *ThemeAwareResource {
	return &ThemeAwareResource{
		LightResource: lightResource,
		DarkResource:  darkResource,
	}
}

// Name returns the name of the resource.
func (r *ThemeAwareResource) Name() string {
	return r.LightResource.Name()
}

func (r *ThemeAwareResource) Content() []byte {
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	if variant == theme.VariantLight {
		return r.LightResource.Content()
	}

	return r.DarkResource.Content()
}
