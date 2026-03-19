// Package ui contains every Fyne screen and shared UI controller for the app.
// All the screens live in this package so they can share the MainUI instance
// without passing it through a pile of constructor parameters.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"languageapp/claude"
	"languageapp/models"
	"languageapp/storage"
)

// MainUI is the central hub that all screens hang off. It owns the shared
// state (active profile, AI client, storage) and handles navigation by
// swapping out the content area.
//
// The pattern here is essentially a lightweight single-page-app router:
// one window, one content slot, screens pushed in and out as needed.
type MainUI struct {
	window         fyne.Window
	store          *storage.Store   // nil if the config dir couldn't be created
	claudeClient   *claude.Client   // nil until the user saves a valid API key
	currentSession *models.Session  // nil between sessions; set by showPractice
	activeProfile  *models.UserProfile // nil until profile is chosen; set by profile screen
	content        *fyne.Container  // the swappable centre area of the window
	statusBar      *widget.Label    // one-line feedback strip along the bottom
	avatarLabel    *widget.Label    // the active user's avatar token (left of status bar)
}

// NewMainUI wires up shared state and returns a ready-to-use controller.
// It attempts to load a previously-saved API key so returning users don't
// have to re-enter it on every launch.
func NewMainUI(w fyne.Window) *MainUI {
	// Try to open the config directory. If it fails for some reason we
	// carry on with store = nil and skip any disk operations gracefully.
	store, err := storage.NewStore()
	if err != nil {
		store = nil
	}

	m := &MainUI{
		window:      w,
		store:       store,
		statusBar:   widget.NewLabel("Welcome! Select or create a profile to get started."),
		avatarLabel: widget.NewLabel(""),
	}

	// If there's a saved API key, build a client right away so the user can
	// go straight from profile selection to practising without an extra step.
	if store != nil {
		if settings, err := store.LoadSettings(); err == nil && settings.APIKey != "" {
			if client, err := claude.NewClient(settings.APIKey); err == nil {
				m.claudeClient = client
			}
		}
	}

	return m
}

// Build creates and returns the root layout for the window. Called once from
// main.go and passed to w.SetContent. The layout is:
//
//	┌──────────────────────────────────────┐
//	│              content area            │  ← screens are swapped in here
//	├──────────────────────────────────────┤
//	│  [avatar]  status bar text           │  ← always visible at the bottom
//	└──────────────────────────────────────┘
func (m *MainUI) Build() fyne.CanvasObject {
	// container.NewMax lets us replace children without rebuilding the layout.
	m.content = container.NewMax()

	// Start on the profile selection screen.
	m.showProfileSelect()

	// Status bar: avatar on the far left, message text filling the rest.
	statusRow := container.NewBorder(nil, nil, m.avatarLabel, nil, m.statusBar)

	return container.NewBorder(
		nil,                                             // no top widget
		container.NewVBox(widget.NewSeparator(), statusRow), // bottom strip
		nil, nil,                                        // no side panels
		m.content,                                       // main body
	)
}

// navigate replaces whatever is currently in the content area with the given screen.
// Calling Refresh() forces Fyne to redraw the container immediately.
func (m *MainUI) navigate(screen fyne.CanvasObject) {
	m.content.Objects = []fyne.CanvasObject{screen}
	m.content.Refresh()
}

// showProfileSelect navigates to the profile picker screen.
// Called on startup and after deleting a profile.
func (m *MainUI) showProfileSelect() {
	m.navigate(NewProfileScreen(m).Build())
}

// showHome navigates to the home / setup screen.
// Called after a profile is selected and when the user clicks "Back to Home"
// from any practice or stats screen.
func (m *MainUI) showHome() {
	m.navigate(NewHomeScreen(m).Build())
}

// showPractice starts a brand-new session and navigates to the practice screen.
// A fresh Session object is created here so scores reset on each visit.
func (m *MainUI) showPractice(lang models.Language, difficulty models.DifficultyLevel) {
	m.currentSession = models.NewSession(lang, difficulty)
	m.navigate(NewPracticeScreen(m, lang, difficulty).Build())
}

// showProgress navigates to the stats / weak-spots screen.
func (m *MainUI) showProgress() {
	m.navigate(NewProgressScreen(m).Build())
}

// showVocab navigates to the personal vocabulary list screen.
func (m *MainUI) showVocab() {
	m.navigate(NewVocabScreen(m).Build())
}

// setStatus updates the one-line message at the bottom of the window.
// Called from various screens to give the user contextual hints.
func (m *MainUI) setStatus(msg string) {
	m.statusBar.SetText(msg)
}

// setAvatar updates the avatar token shown to the left of the status bar.
// Pass an empty string to clear it (e.g., after a profile is deleted).
func (m *MainUI) setAvatar(label string) {
	m.avatarLabel.SetText(label)
}
