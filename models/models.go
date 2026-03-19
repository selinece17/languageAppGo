// Package models defines the data types shared across every layer of the app —
// the UI, the storage layer, and the AI client all import from here.
// Keeping these in one place avoids circular imports and makes it easy to see
// what the app actually tracks.
package models

import "time"

// Language pairs a short BCP-47 code with a display name.
// The code is used when constructing AI prompts; the name is shown in the UI.
type Language struct {
	Code string // e.g. "es", "ja"
	Name string // e.g. "Spanish", "Japanese"
}

// SupportedLanguages is the full list of languages the user can practise.
// Adding a new one here is all that's needed — the UI drop-down is built
// dynamically from this slice.
var SupportedLanguages = []Language{
	{Code: "es", Name: "Spanish"},
	{Code: "fr", Name: "French"},
	{Code: "de", Name: "German"},
	{Code: "it", Name: "Italian"},
	{Code: "pt", Name: "Portuguese"},
	{Code: "ja", Name: "Japanese"},
	{Code: "zh", Name: "Chinese"},
	{Code: "ko", Name: "Korean"},
	{Code: "ar", Name: "Arabic"},
	{Code: "ru", Name: "Russian"},
}

// AvailableAvatars are the little text-art tokens users can pick as a
// profile picture. We use bracketed ASCII rather than emoji so they
// render consistently across every OS and font.
var AvailableAvatars = []string{
	"[*]", "[@]", "[#]", "[!]", "[+]", "[~]",
	"[A]", "[B]", "[C]", "[D]", "[E]", "[F]",
}

// DifficultyLevel controls how complex the AI-generated sentences are.
// It's a string type so we can print it directly in the UI without a lookup table.
type DifficultyLevel string

const (
	DifficultyBeginner     DifficultyLevel = "Beginner"
	DifficultyIntermediate DifficultyLevel = "Intermediate"
	DifficultyAdvanced     DifficultyLevel = "Advanced"
)

// AttemptResult captures everything about one translation attempt so we can
// show history, compute accuracy, and surface weak spots later.
type AttemptResult struct {
	Prompt     string          // the English sentence the user was asked to translate
	UserAnswer string          // what the user actually typed
	IsCorrect  bool            // did the AI judge it correct?
	Feedback   string          // the AI's explanation (stored for the progress screen)
	Language   string          // target language name, e.g. "Spanish"
	Timestamp  time.Time       // when the attempt happened
}

// VocabEntry records a word or phrase the user got wrong so they can
// review it on the vocabulary screen. ReviewCount tracks how many times
// the same English prompt has tripped them up.
type VocabEntry struct {
	English       string    // the original English sentence
	WrongAnswer   string    // what the user typed (for comparison)
	CorrectAnswer string    // what the AI said the right answer was
	Language      string    // target language, e.g. "French"
	AddedAt       time.Time // first time this entry was created
	ReviewCount   int       // incremented each time the same prompt is missed again
}

// WeakSpot aggregates repeated failures on a particular topic so the
// progress screen can highlight what the user should study more.
type WeakSpot struct {
	Topic    string    // a short description of the problem area
	Language string    // which language this weak spot belongs to
	Count    int       // total number of failures in this area
	LastSeen time.Time // when the most recent failure happened
}

// UserProfile stores the identity info for one learner. Multiple profiles
// can exist so a single device can be shared by a family or classroom.
type UserProfile struct {
	Name      string    // display name, also used as the directory key on disk
	Avatar    string    // one of the AvailableAvatars tokens
	CreatedAt time.Time // when the profile was first created
}

// Session tracks the in-memory state of a single practice run. It is NOT
// persisted to disk — it lives only for the duration of one sitting.
// Permanent history is written to storage by RecordAttempt instead.
type Session struct {
	Language   Language
	Difficulty DifficultyLevel
	History    []AttemptResult // all attempts made so far this session
	Score      int             // number of correct answers
	Total      int             // total number of questions answered
}

// NewSession returns an empty Session ready to receive attempts.
func NewSession(lang Language, difficulty DifficultyLevel) *Session {
	return &Session{
		Language:   lang,
		Difficulty: difficulty,
		History:    []AttemptResult{},
	}
}

// AddResult appends an attempt to the session and updates the running totals.
// Call this every time the user submits an answer.
func (s *Session) AddResult(result AttemptResult) {
	s.History = append(s.History, result)
	s.Total++
	if result.IsCorrect {
		s.Score++
	}
}

// Accuracy returns the percentage of correct answers as a value in [0, 100].
// Returns 0 if no attempts have been made yet (avoids a divide-by-zero).
func (s *Session) Accuracy() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Score) / float64(s.Total) * 100
}
