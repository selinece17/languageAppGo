// profile_screen.go — the first thing the user sees when the app starts.
//
// It lists existing profiles so returning users can jump straight in, and
// provides a form to create a new one with a name + avatar token.
// Profiles can also be deleted here (with a confirmation dialog).
package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"languageapp/models"
)

// ProfileScreen is the landing page for the app.
// It holds just a pointer back to MainUI so it can read from storage
// and push navigation events.
type ProfileScreen struct {
	main *MainUI
}

// NewProfileScreen creates a ProfileScreen linked to the given controller.
func NewProfileScreen(m *MainUI) *ProfileScreen {
	return &ProfileScreen{main: m}
}

// Build assembles the profile list and creation form.
// The whole thing is wrapped in a VScroll so it works on small screens even
// if there are many profiles or a longer avatar grid.
func (p *ProfileScreen) Build() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("AI Language Tutor", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabelWithStyle("Who is learning today?", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	profiles := p.loadProfiles()

	// ── Existing profiles list ────────────────────────────────────────────────

	profileList := container.NewVBox()

	if len(profiles) == 0 {
		// Give the user a nudge instead of showing an empty box.
		profileList.Add(widget.NewLabelWithStyle(
			"No profiles yet - create one below!",
			fyne.TextAlignCenter,
			fyne.TextStyle{Italic: true},
		))
	}

	for _, profile := range profiles {
		// Capture the loop variable so each closure references its own profile.
		prof := profile

		// The main button selects this profile and navigates to the home screen.
		btn := widget.NewButton(
			fmt.Sprintf("%s  %s", prof.Avatar, prof.Name),
			func() { p.selectProfile(prof) },
		)
		btn.Importance = widget.MediumImportance

		// Small delete button on the right — confirm before doing anything destructive.
		delBtn := widget.NewButton("X", func() {
			d := dialog.NewConfirm(
				"Delete Profile",
				fmt.Sprintf("Delete '%s' and all their progress?", prof.Name),
				func(confirmed bool) {
					if confirmed {
						p.deleteProfile(prof.Name)
					}
				},
				p.main.window,
			)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
		})

		// Profile button fills the row; delete button is pinned to the right edge.
		row := container.NewBorder(nil, nil, nil, delBtn, btn)
		profileList.Add(container.NewPadded(row))
	}

	// ── Create new profile form ───────────────────────────────────────────────

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your name...")

	// Track the currently selected avatar in a local variable because the
	// Select widget doesn't exist — we use custom buttons instead.
	selectedAvatar := models.AvailableAvatars[0]
	avatarLabel := widget.NewLabel(selectedAvatar)
	avatarLabel.TextStyle = fyne.TextStyle{Bold: true}

	// One button per avatar; tapping one updates both selectedAvatar and the preview label.
	avatarBtns := container.NewGridWithColumns(6)
	for _, av := range models.AvailableAvatars {
		a := av // capture loop variable
		avatarBtns.Add(widget.NewButton(a, func() {
			selectedAvatar = a
			avatarLabel.SetText(a)
		}))
	}

	createBtn := widget.NewButton("Create Profile", func() {
		// Pass the current value of selectedAvatar — by the time this fires,
		// the user will have picked one (or it defaults to the first option).
		p.handleCreate(strings.TrimSpace(nameEntry.Text), selectedAvatar)
	})
	createBtn.Importance = widget.HighImportance

	createCard := widget.NewCard("Create New Profile", "", container.NewVBox(
		container.NewGridWithColumns(2, widget.NewLabel("Name:"), nameEntry),
		widget.NewLabel("Pick your avatar:"),
		avatarBtns,
		container.NewHBox(widget.NewLabel("Selected:"), avatarLabel),
		container.NewPadded(createBtn),
	))

	content := container.NewVBox(
		container.NewPadded(title),
		container.NewPadded(subtitle),
		widget.NewSeparator(),
		container.NewPadded(widget.NewCard("Select Profile", "", profileList)),
		container.NewPadded(createCard),
	)

	return container.NewVScroll(content)
}

// loadProfiles is a small helper that returns an empty slice (rather than
// propagating errors) so Build() stays readable.
func (p *ProfileScreen) loadProfiles() []models.UserProfile {
	if p.main.store == nil {
		return nil
	}
	profiles, err := p.main.store.LoadProfiles()
	if err != nil {
		return nil
	}
	return profiles
}

// selectProfile makes the given profile the active one, updates the status bar
// avatar, and navigates to the home screen.
func (p *ProfileScreen) selectProfile(profile models.UserProfile) {
	p.main.activeProfile = &profile
	p.main.setAvatar(profile.Avatar)
	p.main.setStatus(fmt.Sprintf("Welcome back, %s! Choose a language to practice.", profile.Name))
	p.main.showHome()
}

// handleCreate validates the form inputs and then either saves the profile and
// navigates forward, or shows an informative error dialog.
func (p *ProfileScreen) handleCreate(name string, avatar string) {
	if name == "" {
		d := dialog.NewInformation("Error", "Please enter a name for your profile.", p.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// Cap the name length to keep filenames and display strings sane.
	if len(name) > 20 {
		d := dialog.NewInformation("Error", "Name must be 20 characters or fewer.", p.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	profile := models.UserProfile{
		Name:      name,
		Avatar:    avatar,
		CreatedAt: time.Now(),
	}

	if p.main.store != nil {
		if err := p.main.store.AddProfile(profile); err != nil {
			// AddProfile returns a descriptive error (e.g. "profile already exists")
			// so we can show it directly.
			d := dialog.NewInformation("Error", err.Error(), p.main.window)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
	}

	// Treat the freshly created profile as the active one — no need to go back
	// and click on it in the list.
	p.selectProfile(profile)
}

// deleteProfile removes a profile from storage and refreshes the screen.
// If the deleted profile was the active one, we clear the active state so
// the rest of the app doesn't try to write data to a non-existent user dir.
func (p *ProfileScreen) deleteProfile(name string) {
	if p.main.store != nil {
		if err := p.main.store.DeleteProfile(name); err != nil {
			d := dialog.NewInformation("Error", err.Error(), p.main.window)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
	}

	// Clear active profile if we just deleted whoever was logged in.
	if p.main.activeProfile != nil && strings.EqualFold(p.main.activeProfile.Name, name) {
		p.main.activeProfile = nil
		p.main.setAvatar("")
	}

	// Navigate back to the (now updated) profile list.
	p.main.showProfileSelect()
}
