// vocab_screen.go — shows the user every word/phrase they've got wrong
// so they can review and drill their weak spots.
//
// Entries are added automatically when the practice screen records an
// incorrect answer. Users can remove individual entries or wipe the
// whole list from here.
package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"languageapp/models"
)

// VocabScreen displays the personal vocabulary list for the active profile.
type VocabScreen struct {
	main *MainUI
}

// NewVocabScreen creates a VocabScreen linked to the given MainUI.
func NewVocabScreen(m *MainUI) *VocabScreen {
	return &VocabScreen{main: m}
}

// Build loads the vocabulary list and assembles the layout.
// Three possible states:
//   - no active profile / no store  → shows a brief message
//   - vocab list is empty           → shows an encouraging placeholder
//   - vocab list has entries        → shows the full card list with controls
func (v *VocabScreen) Build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("My Vocabulary List", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	backBtn := widget.NewButton("Back to Home", func() { v.main.showHome() })

	// Guard: should always have a profile here, but fail gracefully just in case.
	if v.main.activeProfile == nil || v.main.store == nil {
		return container.NewVBox(
			container.NewPadded(title),
			widget.NewLabel("No profile active."),
			container.NewPadded(backBtn),
		)
	}

	vocab, err := v.main.store.LoadVocab(v.main.activeProfile.Name)
	if err != nil {
		return container.NewVBox(
			container.NewPadded(title),
			widget.NewLabel(fmt.Sprintf("Error loading vocab: %v", err)),
			container.NewPadded(backBtn),
		)
	}

	// ── Empty state ───────────────────────────────────────────────────────────

	if len(vocab) == 0 {
		emptyLabel := widget.NewLabelWithStyle(
			"Your vocab list is empty!\nWhen you get a question wrong, it will appear here.",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		)
		return container.NewVBox(
			container.NewPadded(title),
			widget.NewSeparator(),
			container.NewCenter(container.NewPadded(emptyLabel)),
			container.NewPadded(backBtn),
		)
	}

	// ── Vocab list ────────────────────────────────────────────────────────────

	// Quick count so the user knows at a glance how much reviewing is waiting.
	subtitle := widget.NewLabel(fmt.Sprintf("%d word(s) to review", len(vocab)))

	list := container.NewVBox()
	for _, entry := range vocab {
		e := entry // capture loop variable before passing to the card builder
		list.Add(v.buildEntryCard(e))
	}

	// Bulk-clear option with a confirmation dialog — destructive actions should
	// always ask first.
	clearBtn := widget.NewButton("Clear Vocab List", func() {
		d := dialog.NewConfirm("Clear Vocabulary", "Delete all vocabulary entries?",
			func(confirmed bool) {
				if confirmed && v.main.store != nil {
					// Replace the slice with an empty one; SaveVocab writes it to disk.
					_ = v.main.store.SaveVocab(v.main.activeProfile.Name, []models.VocabEntry{})
					// Refresh the screen so the empty state is shown immediately.
					v.main.showVocab()
				}
			},
			v.main.window,
		)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
	})
	clearBtn.Importance = widget.DangerImportance

	content := container.NewVBox(
		container.NewPadded(title),
		container.NewPadded(subtitle),
		widget.NewSeparator(),
		container.NewPadded(list),
		widget.NewSeparator(),
		container.NewPadded(clearBtn),
		container.NewPadded(backBtn),
	)

	return container.NewVScroll(content)
}

// buildEntryCard creates a self-contained card for one vocabulary entry.
// Each card shows:
//   - the original English sentence (bold)
//   - what the user typed (their wrong answer)
//   - what the correct answer actually was
//   - metadata: language, date added, review count
//   - a Remove button to delete just this entry
func (v *VocabScreen) buildEntryCard(entry models.VocabEntry) fyne.CanvasObject {
	// English sentence — bold because it's the "question" the user should
	// be able to translate on demand.
	englishLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("EN: %s", entry.English),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	englishLabel.Wrapping = fyne.TextWrapWord

	// Show the user's original wrong answer so they can see where they went
	// off track — useful for spotting patterns in their mistakes.
	wrongLabel := widget.NewLabel(fmt.Sprintf("Your answer:  %s", entry.WrongAnswer))
	wrongLabel.Wrapping = fyne.TextWrapWord

	correctLabel := widget.NewLabel(fmt.Sprintf("Correct:      %s", entry.CorrectAnswer))
	correctLabel.Wrapping = fyne.TextWrapWord

	// Small italic line with housekeeping info.
	metaLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("%s  |  Added %s  |  Reviewed %d time(s)",
			entry.Language, entry.AddedAt.Format("Jan 2"), entry.ReviewCount),
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	// Per-entry remove button — deletes just this one word/phrase and immediately
	// refreshes the screen so the card disappears.
	delBtn := widget.NewButton("Remove", func() {
		if v.main.store != nil {
			_ = v.main.store.DeleteVocabEntry(v.main.activeProfile.Name, entry.English, entry.Language)
			v.main.showVocab()
		}
	})

	return widget.NewCard("", "", container.NewVBox(
		englishLabel,
		wrongLabel,
		correctLabel,
		metaLabel,
		container.NewPadded(delBtn),
	))
}
