// theme.go — a thin custom Fyne theme that swaps in a teal/cyan palette while
// leaving every other design decision (fonts, spacing, icons) to the default.
//
// Why teal? It reads well in both light and dark variants, feels calm and
// focused (good for a study app), and stands out clearly from the neutral
// greys Fyne uses by default.
package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// appTheme satisfies fyne.Theme with a minimal two-colour override.
// Everything not listed in Color() is delegated to theme.DefaultTheme().
type appTheme struct{}

// AppTheme returns the singleton theme value to pass to app.Settings().SetTheme().
// We return the interface rather than a concrete pointer so callers don't need
// to import this package's internal type.
func AppTheme() fyne.Theme {
	return &appTheme{}
}

// Color overrides the two colours that most affect the app's look:
//   - Primary (used for high-importance buttons, selected items)  → teal
//   - Focus  (border colour when a widget is focused)             → cyan
//
// All other colour names fall through to the default theme so dark mode,
// disabled states, etc. still work without us having to enumerate every case.
func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		// Teal (#009688) — Material Design "Teal 500"
		return color.NRGBA{R: 0x00, G: 0x96, B: 0x88, A: 0xFF}
	case theme.ColorNameFocus:
		// Cyan (#00BFD8) — slightly brighter so focused inputs are easy to spot
		return color.NRGBA{R: 0x00, G: 0xBF, B: 0xD8, A: 0xFF}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font delegates entirely to the default theme. We don't bundle a custom font
// so the app stays lightweight and renders with the OS system font.
func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon delegates entirely to the default theme.
func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size delegates entirely to the default theme so padding and text sizes
// respect the user's system accessibility settings.
func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
