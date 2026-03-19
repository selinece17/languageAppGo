// home_screen.go — the setup screen a user sees immediately after picking
// (or creating) their profile. From here they configure their API key,
// choose a language + difficulty, and jump into a practice session.
package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"languageapp/claude"
	"languageapp/models"
	"languageapp/storage"
)

// HomeScreen is the screen shown after a profile is selected.
// It holds a reference to MainUI so it can read shared state (active profile,
// store, etc.) and trigger navigation.
type HomeScreen struct {
	main *MainUI
}

// NewHomeScreen constructs a HomeScreen linked to the given MainUI controller.
func NewHomeScreen(m *MainUI) *HomeScreen {
	return &HomeScreen{main: m}
}

// Build assembles all the widgets and returns the finished layout.
// It's called fresh every time the home screen is navigated to, so the
// displayed name/avatar always reflects whoever is currently logged in.
func (h *HomeScreen) Build() fyne.CanvasObject {
	// Pull display values out of the active profile (or use safe defaults if
	// somehow no profile is set — shouldn't happen in normal flow but good
	// to be defensive).
	userName := "User"
	avatarStr := "[*]"
	if h.main.activeProfile != nil {
		userName = h.main.activeProfile.Name
		avatarStr = h.main.activeProfile.Avatar
	}

	// Greeting at the top of the card.
	title := widget.NewLabelWithStyle(
		fmt.Sprintf("%s  Hello, %s!", avatarStr, userName),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	subtitle := widget.NewLabelWithStyle(
		"Powered by Google Gemini AI",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// ── API Key section ───────────────────────────────────────────────────────

	apiKeyLabel := widget.NewLabel("Google Gemini API Key:")

	// PasswordEntry masks the key so it's not visible over someone's shoulder.
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("AIza...")

	// Pre-fill the field if we already have a saved key — saves re-entering it
	// every launch.
	if h.main.store != nil {
		if settings, err := h.main.store.LoadSettings(); err == nil {
			apiKeyEntry.SetText(settings.APIKey)
		}
	}

	saveKeyBtn := widget.NewButton("Save Key", func() {
		h.handleSaveKey(apiKeyEntry.Text)
	})
	saveKeyBtn.Importance = widget.HighImportance

	// Border layout: label on the left, button on the right, entry fills middle.
	apiKeyRow := container.NewBorder(nil, nil, apiKeyLabel, saveKeyBtn, apiKeyEntry)

	// ── Language selector ─────────────────────────────────────────────────────

	// Build the drop-down options from the canonical SupportedLanguages slice
	// so this UI and the models package are always in sync.
	langNames := make([]string, len(models.SupportedLanguages))
	for i, l := range models.SupportedLanguages {
		langNames[i] = l.Name
	}
	langSelect := widget.NewSelect(langNames, nil)
	langSelect.SetSelected(langNames[0]) // default to Spanish

	// ── Difficulty selector ───────────────────────────────────────────────────

	diffSelect := widget.NewSelect([]string{
		string(models.DifficultyBeginner),
		string(models.DifficultyIntermediate),
		string(models.DifficultyAdvanced),
	}, nil)
	diffSelect.SetSelected(string(models.DifficultyBeginner))

	// Restore whatever the user picked last time so they don't have to
	// reselect their preferred language every session.
	if h.main.store != nil {
		if settings, err := h.main.store.LoadSettings(); err == nil {
			for i, l := range models.SupportedLanguages {
				if l.Code == settings.DefaultLanguage {
					langSelect.SetSelected(langNames[i])
					break
				}
			}
			if settings.DefaultDifficulty != "" {
				diffSelect.SetSelected(settings.DefaultDifficulty)
			}
		}
	}

	// ── Navigation buttons ────────────────────────────────────────────────────

	startBtn := widget.NewButton("Start Practicing", func() {
		h.handleStart(langSelect.SelectedIndex(), diffSelect.Selected)
	})
	startBtn.Importance = widget.HighImportance

	vocabBtn := widget.NewButton("My Vocabulary List", func() { h.main.showVocab() })
	progressBtn := widget.NewButton("View My Progress", func() { h.main.showProgress() })
	switchBtn := widget.NewButton("Switch Profile", func() { h.main.showProfileSelect() })

	// Stack everything vertically inside a card for the centered look.
	form := container.NewVBox(
		container.NewPadded(container.NewVBox(title, subtitle, widget.NewSeparator())),
		container.NewPadded(apiKeyRow),
		container.NewPadded(container.NewGridWithColumns(2, widget.NewLabel("Language:"), langSelect)),
		container.NewPadded(container.NewGridWithColumns(2, widget.NewLabel("Difficulty:"), diffSelect)),
		container.NewPadded(startBtn),
		container.NewPadded(vocabBtn),
		container.NewPadded(progressBtn),
		container.NewPadded(switchBtn),
	)

	return container.NewCenter(widget.NewCard("", "", form))
}

// handleSaveKey validates the entered key, creates a Gemini client with it,
// and persists it to disk so the user doesn't have to enter it again next time.
func (h *HomeScreen) handleSaveKey(key string) {
	key = strings.TrimSpace(key)

	if key == "" {
		d := dialog.NewInformation("Error", "API key cannot be empty.", h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// Google AI Studio keys always start with "AIza" — catch obvious mistakes
	// before wasting a network round-trip.
	if !strings.HasPrefix(key, "AIza") {
		d := dialog.NewInformation("Invalid Key Format",
			"Google AI Studio keys start with 'AIza'. Check aistudio.google.com.",
			h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// NewClient does a light validation pass (non-empty); it won't hit the network
	// until Send is called during a practice session.
	client, err := claude.NewClient(key)
	if err != nil {
		d := dialog.NewInformation("Error", err.Error(), h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// Persist the key so it survives app restarts.
	if h.main.store != nil {
		settings, _ := h.main.store.LoadSettings()
		if settings == nil {
			settings = &storage.Settings{}
		}
		settings.APIKey = key
		_ = h.main.store.SaveSettings(settings)
	}

	// Store the live client so the practice screen can use it immediately.
	h.main.claudeClient = client
	h.main.setStatus("API key saved! Choose a language and click Start.")

	d := dialog.NewInformation("Success", "API key saved successfully!", h.main.window)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

// handleStart validates that everything needed for a session is in place,
// saves the current selections as defaults, then navigates to the practice screen.
func (h *HomeScreen) handleStart(langIndex int, difficultyStr string) {
	// Need a working API client before we can fetch questions.
	if h.main.claudeClient == nil {
		d := dialog.NewInformation("No API Key", "Please save your Gemini API key first.", h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// Shouldn't be possible to reach this screen without a profile, but check anyway.
	if h.main.activeProfile == nil {
		d := dialog.NewInformation("No Profile",
			"No profile selected. Please go back and select a profile.",
			h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// Guard against an invalid index — could happen if SupportedLanguages is empty.
	if langIndex < 0 || langIndex >= len(models.SupportedLanguages) {
		d := dialog.NewInformation("Error", "Please select a language.", h.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	selectedLang := models.SupportedLanguages[langIndex]
	difficulty := models.DifficultyLevel(difficultyStr)

	// Save selections as the new defaults so next launch feels consistent.
	if h.main.store != nil {
		if settings, err := h.main.store.LoadSettings(); err == nil {
			settings.DefaultLanguage = selectedLang.Code
			settings.DefaultDifficulty = difficultyStr
			_ = h.main.store.SaveSettings(settings)
		}
	}

	h.main.showPractice(selectedLang, difficulty)
}
