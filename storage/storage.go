// Package storage handles everything that touches the filesystem.
// Each user gets their own subdirectory so profiles stay isolated and
// deleting one doesn't affect any others.
//
// Directory layout under ~/.config/languageapp/:
//
//	settings.json          — global prefs (API key, last-used language/difficulty)
//	profiles.json          — ordered list of all UserProfile records
//	users/<name>/
//	    progress.json      — attempt history, weak spots, calendar of practice days
//	    vocab.json         — personal dictionary of words/phrases got wrong
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"languageapp/models"
)

// Settings holds preferences that apply globally, regardless of which profile
// is active. Persisted across launches so the user doesn't re-enter their key.
type Settings struct {
	APIKey            string `json:"api_key"`
	DefaultLanguage   string `json:"default_language"`   // BCP-47 code, e.g. "es"
	DefaultDifficulty string `json:"default_difficulty"` // "Beginner" / "Intermediate" / "Advanced"
	ActiveProfile     string `json:"active_profile"`     // name of whichever profile was last used
}

// Progress bundles everything we store long-term for a single user.
// It's intentionally flat — no nested databases — so the JSON files are
// human-readable and easy to back up.
type Progress struct {
	AllAttempts  []models.AttemptResult `json:"all_attempts"`
	WeakSpots    []models.WeakSpot      `json:"weak_spots"`
	PracticeDays []string               `json:"practice_days"` // "YYYY-MM-DD" strings, one per calendar day with at least one attempt
}

// Store is the single entry point for all disk I/O. Create it with NewStore
// and pass it around; don't create multiple instances or you risk races on
// the JSON files.
type Store struct {
	dir string // absolute path to ~/.config/languageapp (or "." as a fallback)
}

// NewStore locates (or creates) the app's config directory and returns a Store
// rooted there. The directory is created with 0755 permissions if absent.
func NewStore() (*Store, error) {
	// os.UserConfigDir gives us the right path for each OS:
	//   Linux/macOS → ~/.config
	//   Windows     → %AppData%
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Unusual, but possible on headless systems. Fall back to CWD.
		configDir = "."
	}

	dir := filepath.Join(configDir, "languageapp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Store{dir: dir}, nil
}

// ─── Settings ─────────────────────────────────────────────────────────────────

// LoadSettings reads settings.json. If the file doesn't exist yet (first launch),
// it returns sensible defaults rather than an error.
func (s *Store) LoadSettings() (*Settings, error) {
	path := filepath.Join(s.dir, "settings.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// First run — hand back defaults so the rest of the app doesn't have
		// to special-case a nil settings pointer.
		return &Settings{
			DefaultLanguage:   "es",
			DefaultDifficulty: string(models.DifficultyBeginner),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}
	return &settings, nil
}

// SaveSettings writes settings to settings.json, creating or overwriting it.
func (s *Store) SaveSettings(settings *Settings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// MarshalIndent keeps the file readable if someone opens it in a text editor.
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize settings: %w", err)
	}

	return os.WriteFile(filepath.Join(s.dir, "settings.json"), data, 0644)
}

// ─── User Profiles ────────────────────────────────────────────────────────────

// LoadProfiles reads the full list of profiles from profiles.json.
// Returns an empty slice (not an error) when no profiles have been created yet.
func (s *Store) LoadProfiles() ([]models.UserProfile, error) {
	path := filepath.Join(s.dir, "profiles.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []models.UserProfile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles: %w", err)
	}

	var profiles []models.UserProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles: %w", err)
	}
	return profiles, nil
}

// SaveProfiles overwrites profiles.json with the given slice.
// This is the underlying write operation; callers that only want to add or
// remove one profile should use AddProfile / DeleteProfile instead.
func (s *Store) SaveProfiles(profiles []models.UserProfile) error {
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize profiles: %w", err)
	}
	return os.WriteFile(filepath.Join(s.dir, "profiles.json"), data, 0644)
}

// AddProfile appends a new profile and creates its data directory.
// Names are compared case-insensitively to prevent confusing duplicates like
// "Alice" and "alice" sharing different data dirs.
func (s *Store) AddProfile(profile models.UserProfile) error {
	profiles, err := s.LoadProfiles()
	if err != nil {
		return err
	}

	// Reject duplicates up front rather than silently clobbering data.
	for _, p := range profiles {
		if strings.EqualFold(p.Name, profile.Name) {
			return fmt.Errorf("a profile named '%s' already exists", profile.Name)
		}
	}

	profiles = append(profiles, profile)

	// Make sure the user's data directory exists before we write profiles,
	// so there's never a profile record without a matching directory.
	if err := os.MkdirAll(s.userDir(profile.Name), 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	return s.SaveProfiles(profiles)
}

// DeleteProfile removes the profile record AND wipes the user's data directory.
// This is intentionally destructive — we show a confirmation dialog in the UI
// before calling this.
func (s *Store) DeleteProfile(name string) error {
	profiles, err := s.LoadProfiles()
	if err != nil {
		return err
	}

	// Build a new slice that excludes the deleted profile.
	// We use a zero-length slice backed by the same array to avoid an allocation.
	filtered := profiles[:0]
	for _, p := range profiles {
		if !strings.EqualFold(p.Name, name) {
			filtered = append(filtered, p)
		}
	}

	if err := s.SaveProfiles(filtered); err != nil {
		return err
	}

	// Blow away the directory tree — vocab, progress, everything.
	return os.RemoveAll(s.userDir(name))
}

// userDir returns the filesystem path for a given user's private data.
// The name is lowercased and spaces are replaced with underscores so it's
// a valid directory name on all supported OSes.
func (s *Store) userDir(name string) string {
	safe := strings.ReplaceAll(strings.ToLower(name), " ", "_")
	return filepath.Join(s.dir, "users", safe)
}

// ─── Progress ─────────────────────────────────────────────────────────────────

// LoadProgress reads progress.json for the given user.
// Returns a zeroed-out Progress struct (empty slices, not nil) if the file
// doesn't exist, so callers can append to it without nil checks.
func (s *Store) LoadProgress(userName string) (*Progress, error) {
	path := filepath.Join(s.userDir(userName), "progress.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Progress{
			AllAttempts:  []models.AttemptResult{},
			WeakSpots:    []models.WeakSpot{},
			PracticeDays: []string{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read progress: %w", err)
	}

	var progress Progress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to parse progress: %w", err)
	}
	return &progress, nil
}

// SaveProgress writes the full Progress struct to disk for the given user.
// It creates the user's directory if it doesn't exist yet.
func (s *Store) SaveProgress(userName string, progress *Progress) error {
	if progress == nil {
		return fmt.Errorf("progress cannot be nil")
	}

	// Make sure the directory is there — this is a no-op if it already exists.
	if err := os.MkdirAll(s.userDir(userName), 0755); err != nil {
		return fmt.Errorf("failed to create user dir: %w", err)
	}

	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize progress: %w", err)
	}

	return os.WriteFile(filepath.Join(s.userDir(userName), "progress.json"), data, 0644)
}

// RecordAttempt is the main write path after each practice question.
// It appends the attempt, stamps today as a practice day (de-duped), and
// updates weak-spot counters if the answer was wrong.
func (s *Store) RecordAttempt(userName string, result models.AttemptResult) error {
	progress, err := s.LoadProgress(userName)
	if err != nil {
		return err
	}

	// Append the raw attempt record — we keep everything so the history
	// screen can show accurate stats even after weak spots are cleared.
	progress.AllAttempts = append(progress.AllAttempts, result)

	// Mark today as a practice day. We store dates as strings so they're
	// trivially readable in the JSON file.
	today := time.Now().Format("2006-01-02")
	alreadyLogged := false
	for _, d := range progress.PracticeDays {
		if d == today {
			alreadyLogged = true
			break
		}
	}
	if !alreadyLogged {
		progress.PracticeDays = append(progress.PracticeDays, today)
	}

	// Only update weak spots for wrong answers; correct ones are already tracked
	// by the overall accuracy stat.
	if !result.IsCorrect {
		s.updateWeakSpots(progress, result)
	}

	return s.SaveProgress(userName, progress)
}

// updateWeakSpots increments the counter for an existing weak-spot entry, or
// creates a new one if this is the first failure in that area. Currently we
// group all failures for a language under one topic string; a future version
// could use the AI feedback to categorise by grammar rule.
func (s *Store) updateWeakSpots(progress *Progress, result models.AttemptResult) {
	// Topic format: "<Language> translation" — simple but descriptive enough for now.
	topic := result.Language + " translation"

	for i, ws := range progress.WeakSpots {
		if ws.Topic == topic && ws.Language == result.Language {
			// Already tracking this area — just bump the counter.
			progress.WeakSpots[i].Count++
			progress.WeakSpots[i].LastSeen = time.Now()
			return
		}
	}

	// First failure in this area — add a fresh entry.
	progress.WeakSpots = append(progress.WeakSpots, models.WeakSpot{
		Topic:    topic,
		Language: result.Language,
		Count:    1,
		LastSeen: time.Now(),
	})
}

// ─── Vocabulary ───────────────────────────────────────────────────────────────

// LoadVocab reads the personal vocabulary list for the given user.
// Returns an empty slice if no vocab file exists yet.
func (s *Store) LoadVocab(userName string) ([]models.VocabEntry, error) {
	path := filepath.Join(s.userDir(userName), "vocab.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []models.VocabEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read vocab: %w", err)
	}

	var vocab []models.VocabEntry
	if err := json.Unmarshal(data, &vocab); err != nil {
		return nil, fmt.Errorf("failed to parse vocab: %w", err)
	}
	return vocab, nil
}

// SaveVocab writes the full vocabulary slice to disk, creating the user
// directory if needed.
func (s *Store) SaveVocab(userName string, vocab []models.VocabEntry) error {
	if err := os.MkdirAll(s.userDir(userName), 0755); err != nil {
		return fmt.Errorf("failed to create user dir: %w", err)
	}

	data, err := json.MarshalIndent(vocab, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize vocab: %w", err)
	}

	return os.WriteFile(filepath.Join(s.userDir(userName), "vocab.json"), data, 0644)
}

// AddVocabEntry appends a new word/phrase to the user's vocabulary list.
// If an entry with the same English text already exists for the same language,
// we just increment its review count rather than duplicating it.
func (s *Store) AddVocabEntry(userName string, entry models.VocabEntry) error {
	vocab, err := s.LoadVocab(userName)
	if err != nil {
		return err
	}

	// Case-insensitive comparison so "Hello" and "hello" count as the same entry.
	for i, v := range vocab {
		if strings.EqualFold(v.English, entry.English) && v.Language == entry.Language {
			vocab[i].ReviewCount++
			return s.SaveVocab(userName, vocab)
		}
	}

	// Brand new entry — add it to the end.
	vocab = append(vocab, entry)
	return s.SaveVocab(userName, vocab)
}

// DeleteVocabEntry removes a single entry identified by its English text and
// target language. If the entry isn't found, the list is saved unchanged
// (no error — idempotent delete is fine here).
func (s *Store) DeleteVocabEntry(userName string, english string, language string) error {
	vocab, err := s.LoadVocab(userName)
	if err != nil {
		return err
	}

	// Filter in-place: keep everything that doesn't match.
	filtered := vocab[:0]
	for _, v := range vocab {
		if !(strings.EqualFold(v.English, english) && v.Language == language) {
			filtered = append(filtered, v)
		}
	}

	return s.SaveVocab(userName, filtered)
}
