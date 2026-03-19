// progress_screen.go — shows the user how they're doing over time.
//
// Three sections:
//  1. Overall stats   — total attempts, accuracy, days practised
//  2. Weak spots      — topics with the most wrong answers
//  3. Recent attempts — last 10 questions in reverse chronological order
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"languageapp/storage"
)

// ProgressScreen shows historical stats for the active user profile.
type ProgressScreen struct {
	main *MainUI
}

// NewProgressScreen creates a ProgressScreen linked to the given MainUI.
func NewProgressScreen(m *MainUI) *ProgressScreen {
	return &ProgressScreen{main: m}
}

// Build loads progress data from disk and assembles the full layout.
// Returns a minimal error view if the data can't be read or no profile
// is active (so the screen is always safe to navigate to).
func (p *ProgressScreen) Build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("My Progress", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	backBtn := widget.NewButton("Back to Home", func() { p.main.showHome() })

	// Safety check — shouldn't normally be nil, but better to show a message
	// than panic.
	if p.main.activeProfile == nil || p.main.store == nil {
		return container.NewVBox(
			container.NewPadded(title),
			widget.NewLabel("No profile active."),
			container.NewPadded(backBtn),
		)
	}

	progress, err := p.main.store.LoadProgress(p.main.activeProfile.Name)
	if err != nil {
		return container.NewVBox(
			container.NewPadded(title),
			widget.NewLabel(fmt.Sprintf("Could not load progress: %v", err)),
			container.NewPadded(backBtn),
		)
	}

	// ── Overall stats ─────────────────────────────────────────────────────────

	total := len(progress.AllAttempts)
	correct := 0
	for _, a := range progress.AllAttempts {
		if a.IsCorrect {
			correct++
		}
	}

	// Compute accuracy as a percentage; guard against divide-by-zero.
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(correct) / float64(total) * 100
	}

	statsLabel := widget.NewLabel(fmt.Sprintf(
		"Total Attempts: %d    Correct: %d    Accuracy: %.1f%%    Days Practiced: %d",
		total, correct, accuracy, len(progress.PracticeDays),
	))
	statsCard := widget.NewCard("Overall Statistics", "", statsLabel)

	// ── Weak spots ────────────────────────────────────────────────────────────

	var weakContent fyne.CanvasObject
	if len(progress.WeakSpots) == 0 {
		// Positive message when there's nothing to show — keeps the tone encouraging.
		weakContent = widget.NewLabel("No weak spots yet. Keep practicing!")
	} else {
		weakList := container.NewVBox()
		for _, ws := range progress.WeakSpots {
			weakList.Add(widget.NewLabel(fmt.Sprintf(
				"- %s (%s) — missed %d time(s), last seen %s",
				ws.Topic, ws.Language, ws.Count, ws.LastSeen.Format("Jan 2"),
			)))
		}
		weakContent = weakList
	}
	weakCard := widget.NewCard("Weak Spots", "", weakContent)

	// ── Recent attempts ───────────────────────────────────────────────────────

	recentCard := p.buildRecentAttempts(progress)

	// ── Danger zone ───────────────────────────────────────────────────────────

	// Clearing progress is irreversible, so we give the button red styling to
	// signal that. No confirmation dialog here — the button label is explicit enough.
	clearBtn := widget.NewButton("Clear All Progress", func() {
		if p.main.store != nil {
			// Overwrite with an empty Progress struct to wipe the slate clean.
			_ = p.main.store.SaveProgress(p.main.activeProfile.Name, &storage.Progress{})
			// Refresh the screen so the zeroed stats are visible immediately.
			p.main.showProgress()
		}
	})
	clearBtn.Importance = widget.DangerImportance

	content := container.NewVBox(
		container.NewPadded(title),
		widget.NewSeparator(),
		container.NewPadded(statsCard),
		container.NewPadded(weakCard),
		container.NewPadded(recentCard),
		widget.NewSeparator(),
		container.NewPadded(clearBtn),
		container.NewPadded(backBtn),
	)

	return container.NewVScroll(content)
}

// buildRecentAttempts creates a card showing up to the last 10 attempts in
// reverse order (most recent first) so the user can quickly see how the
// current session went.
func (p *ProgressScreen) buildRecentAttempts(progress *storage.Progress) fyne.CanvasObject {
	attempts := progress.AllAttempts

	// Slice off everything older than the last 10 entries.
	start := 0
	if len(attempts) > 10 {
		start = len(attempts) - 10
	}
	recent := attempts[start:]

	if len(recent) == 0 {
		return widget.NewCard("Recent Attempts", "", widget.NewLabel("No attempts yet."))
	}

	list := container.NewVBox()

	// Iterate in reverse so the newest attempt is at the top of the card.
	for i := len(recent) - 1; i >= 0; i-- {
		a := recent[i]

		// Simple text tag so the user can scan right/wrong at a glance.
		icon := "[correct]"
		if !a.IsCorrect {
			icon = "[wrong]"
		}

		entry := widget.NewLabel(fmt.Sprintf("%s [%s] %s", icon, a.Language, a.Prompt))
		entry.Wrapping = fyne.TextWrapWord
		list.Add(entry)
		list.Add(widget.NewSeparator()) // visual divider between attempts
	}

	return widget.NewCard("Recent Attempts (last 10)", "", list)
}
