// AI Language Learning App — entry point.
//
// Spins up a Fyne window, applies our custom theme, then hands off all
// the real work to the ui package. Keeping main.go thin means we can
// swap the GUI toolkit later without touching business logic.
//
// Authors: [Your Name]
// Course:  [Course Name]
//
// External references used during development:
//   - Fyne v2 docs:          https://docs.fyne.io
//   - Google Gemini API:     https://ai.google.dev
//   - Go standard library:   https://pkg.go.dev/std
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"languageapp/ui"
)

func main() {
	// app.New() initialises Fyne and must be called before any UI objects
	// are created. It reads the OS's preferred dark/light mode automatically.
	a := app.New()

	// Swap in our teal-accented palette instead of Fyne's default grey.
	a.Settings().SetTheme(ui.AppTheme())

	// Main window — give it enough room for the practice screen to breathe
	// without feeling cramped on a typical laptop display.
	w := a.NewWindow("🌍 AI Language Tutor")
	w.Resize(fyne.NewSize(900, 650))
	w.SetFixedSize(false) // users can resize if they want more room

	// NewMainUI wires together navigation, storage, and the AI client.
	// Build() returns the root CanvasObject (a border layout with a status bar).
	mainUI := ui.NewMainUI(w)
	w.SetContent(mainUI.Build())

	// ShowAndRun blocks until the window is closed, then exits cleanly.
	w.ShowAndRun()
}
