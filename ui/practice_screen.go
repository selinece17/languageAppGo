// practice_screen.go — the core learning screen.
//
// Flow:
//  1. Screen loads  → goroutine fires off to Gemini to fetch the first question.
//  2. User types    → hits "Check My Answer".
//  3. Another goroutine sends prompt + answer to Gemini for evaluation.
//  4. Feedback is shown; wrong answers are auto-saved to the vocab list.
//  5. "Next Question" hides feedback and repeats from step 1.
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

// PracticeScreen owns all the widgets for the translation exercise and keeps
// track of the current prompt so it can pass it to the evaluation call.
type PracticeScreen struct {
	main          *MainUI
	language      models.Language
	difficulty    models.DifficultyLevel
	currentPrompt string                   // the English sentence currently displayed
	answerEntry   *widget.Entry
	promptLabel   *widget.Label
	feedbackLabel *widget.Label
	scoreLabel    *widget.Label
	avatarLabel   *widget.Label            // shows a reaction ("[correct!]" / "[wrong!]") after grading
	submitBtn     *widget.Button
	nextBtn       *widget.Button
	loading       *widget.ProgressBarInfinite
}

// NewPracticeScreen creates a PracticeScreen for the given language and difficulty.
// The screen won't start fetching questions until Build() is called.
func NewPracticeScreen(m *MainUI, lang models.Language, difficulty models.DifficultyLevel) *PracticeScreen {
	return &PracticeScreen{
		main:       m,
		language:   lang,
		difficulty: difficulty,
	}
}

// Build lays out all the widgets and kicks off the first question fetch.
// It returns a scrollable container so the feedback card is reachable on
// small screens without the window needing to be resized.
func (p *PracticeScreen) Build() fyne.CanvasObject {
	// Header shows what language/difficulty this session is.
	header := widget.NewLabelWithStyle(
		fmt.Sprintf("%s Practice  -  %s", p.language.Name, p.difficulty),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Avatar reaction label — sits to the left of the score.
	avatarStr := "[*]"
	if p.main.activeProfile != nil {
		avatarStr = p.main.activeProfile.Avatar
	}
	p.avatarLabel = widget.NewLabelWithStyle(avatarStr, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	p.scoreLabel = widget.NewLabelWithStyle("Score: 0 / 0", fyne.TextAlignCenter, fyne.TextStyle{})
	scoreRow := container.NewBorder(nil, nil, p.avatarLabel, nil, p.scoreLabel)

	// ── Prompt card ───────────────────────────────────────────────────────────

	p.promptLabel = widget.NewLabelWithStyle(
		"Loading your first question...",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	// Allow the sentence to wrap so long prompts don't clip.
	p.promptLabel.Wrapping = fyne.TextWrapWord
	promptCard := widget.NewCard("Translate this sentence:", "", p.promptLabel)

	// ── Answer input ──────────────────────────────────────────────────────────

	p.answerEntry = widget.NewMultiLineEntry()
	p.answerEntry.SetPlaceHolder("Type your translation here...")
	p.answerEntry.SetMinRowsVisible(3)

	// ── Feedback card (hidden until an answer is graded) ──────────────────────

	p.feedbackLabel = widget.NewLabel("")
	p.feedbackLabel.Wrapping = fyne.TextWrapWord
	feedbackCard := widget.NewCard("AI Feedback:", "", p.feedbackLabel)
	feedbackCard.Hide() // only shown after grading

	// ── Loading indicator ─────────────────────────────────────────────────────

	p.loading = widget.NewProgressBarInfinite()
	p.loading.Hide()

	// ── Buttons ───────────────────────────────────────────────────────────────

	p.submitBtn = widget.NewButton("Check My Answer", nil)
	p.submitBtn.Importance = widget.HighImportance
	// Assign OnTapped after the card is declared so we can close over feedbackCard.
	p.submitBtn.OnTapped = func() { p.handleSubmit(feedbackCard) }

	p.nextBtn = widget.NewButton("Next Question", nil)
	p.nextBtn.Hide() // shown only after feedback is displayed
	p.nextBtn.OnTapped = func() {
		// Hide the previous answer's feedback before fetching a new question.
		feedbackCard.Hide()
		p.nextBtn.Hide()
		p.loadNextQuestion()
	}

	vocabBtn := widget.NewButton("My Vocab List", func() { p.main.showVocab() })
	backBtn := widget.NewButton("Back to Home", func() { p.main.showHome() })
	bottomRow := container.NewGridWithColumns(2, vocabBtn, backBtn)

	content := container.NewVBox(
		container.NewPadded(header),
		container.NewPadded(scoreRow),
		widget.NewSeparator(),
		container.NewPadded(promptCard),
		container.NewPadded(p.answerEntry),
		container.NewPadded(p.submitBtn),
		container.NewPadded(p.loading),
		container.NewPadded(feedbackCard),
		container.NewPadded(p.nextBtn),
		widget.NewSeparator(),
		container.NewPadded(bottomRow),
	)

	// Kick off the first question in the background so the UI renders immediately
	// and the user sees "Generating question..." rather than a frozen window.
	go p.loadNextQuestion()

	return container.NewVScroll(content)
}

// loadNextQuestion sends a request to Gemini asking for one English sentence
// to translate. Runs in a goroutine; updates widget state when done.
func (p *PracticeScreen) loadNextQuestion() {
	p.setLoading(true)
	p.answerEntry.SetText("") // clear any previous answer
	p.promptLabel.SetText("Generating question...")

	// The system prompt tells the model to act as a question generator for
	// this specific language and difficulty level.
	systemPrompt := fmt.Sprintf(
		`You are a language learning assistant helping a user practice %s.
Give the user one English sentence to translate into %s.
Adjust complexity for a %s level learner.
Respond with ONLY the English sentence, nothing else.`,
		p.language.Name, p.language.Name, p.difficulty)

	response, err := p.main.claudeClient.Send(systemPrompt, "Give me one sentence to translate.")
	p.setLoading(false)

	if err != nil {
		// Surface the error in the prompt label so the user knows what happened
		// without needing a modal dialog.
		p.promptLabel.SetText("Error loading question. Check your API key and internet connection.")
		p.main.setStatus(fmt.Sprintf("Error: %v", err))
		return
	}

	// TrimSpace cleans up any trailing newline the model might add.
	p.currentPrompt = strings.TrimSpace(response)
	p.promptLabel.SetText(p.currentPrompt)
	p.submitBtn.Enable()
	p.main.setStatus(fmt.Sprintf("Translate into %s and press Check My Answer.", p.language.Name))
}

// handleSubmit validates the user's answer and sends it to Gemini for grading.
// Runs the actual API call in a goroutine to keep the UI responsive.
func (p *PracticeScreen) handleSubmit(feedbackCard *widget.Card) {
	answer := strings.TrimSpace(p.answerEntry.Text)

	// Don't let the user submit an empty box.
	if answer == "" {
		d := dialog.NewInformation("Empty Answer", "Please type your translation before submitting.", p.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	// If the question hasn't loaded yet (edge case: user clicks submit very fast),
	// don't send a grading request with an empty prompt.
	if p.currentPrompt == "" {
		d := dialog.NewInformation("No Question", "Please wait for a question to load first.", p.main.window)
		d.Resize(fyne.NewSize(400, 200))
		d.Show()
		return
	}

	p.setLoading(true)

	go func() {
		// Ask the model to grade the answer in a structured format we can parse.
		// The EXACT format request is important — if the model goes off-format
		// the string parsing below will miss the result/feedback/correct fields.
		systemPrompt := fmt.Sprintf(
			`You are a %s language teacher. The student was asked to translate an English sentence into %s.
Evaluate their answer and respond in this EXACT format:

RESULT: correct
FEEDBACK: [brief encouraging feedback]
CORRECT: [the correct translation]

or

RESULT: incorrect
FEEDBACK: [explain the error and give a helpful tip]
CORRECT: [the correct translation]

Always include all three lines.`,
			p.language.Name, p.language.Name)

		userMsg := fmt.Sprintf(
			"English sentence: %s\nStudent's %s translation: %s",
			p.currentPrompt, p.language.Name, answer,
		)

		response, err := p.main.claudeClient.Send(systemPrompt, userMsg)
		p.setLoading(false)

		if err != nil {
			// Show the error inside the feedback card rather than a modal so the
			// user can still see the question and try again.
			p.feedbackLabel.SetText(fmt.Sprintf("Could not get feedback: %v", err))
			feedbackCard.Show()
			p.submitBtn.Enable()
			return
		}

		// ── Parse the structured response ─────────────────────────────────────

		// A case-insensitive contains check is more robust than a strict prefix match
		// in case the model adds a space or extra punctuation.
		isCorrect := strings.Contains(strings.ToLower(response), "result: correct")

		// Pull out just the FEEDBACK line.
		feedbackText := response
		if idx := strings.Index(strings.ToUpper(response), "FEEDBACK:"); idx >= 0 {
			// The feedback ends where CORRECT: begins (or at EOF if CORRECT is missing).
			end := strings.Index(strings.ToUpper(response[idx:]), "\nCORRECT:")
			if end >= 0 {
				feedbackText = strings.TrimSpace(response[idx+9 : idx+end])
			} else {
				feedbackText = strings.TrimSpace(response[idx+9:])
			}
		}

		// Pull out the CORRECT translation line.
		correctTranslation := ""
		if idx := strings.Index(strings.ToUpper(response), "CORRECT:"); idx >= 0 {
			correctTranslation = strings.TrimSpace(response[idx+8:])
			// Only keep the first line — the model sometimes adds a blank line after it.
			if nl := strings.Index(correctTranslation, "\n"); nl >= 0 {
				correctTranslation = strings.TrimSpace(correctTranslation[:nl])
			}
		}

		// ── Update the UI ─────────────────────────────────────────────────────

		// Swap the avatar label to show a quick reaction.
		if p.main.activeProfile != nil {
			if isCorrect {
				p.avatarLabel.SetText("(correct!)")
			} else {
				p.avatarLabel.SetText("(wrong!)")
			}
		}

		// Prepend a [CORRECT] / [WRONG] tag so the result is obvious even
		// before the user reads the full feedback paragraph.
		resultTag := "[WRONG]"
		if isCorrect {
			resultTag = "[CORRECT]"
		}
		p.feedbackLabel.SetText(fmt.Sprintf("%s  %s", resultTag, feedbackText))
		feedbackCard.Show()

		// ── Record the attempt ────────────────────────────────────────────────

		result := models.AttemptResult{
			Prompt:     p.currentPrompt,
			UserAnswer: answer,
			IsCorrect:  isCorrect,
			Feedback:   feedbackText,
			Language:   p.language.Name,
			Timestamp:  time.Now(),
		}

		// Update the live session score so the counter at the top of the screen
		// reflects the current run immediately.
		if p.main.currentSession != nil {
			p.main.currentSession.AddResult(result)
			s := p.main.currentSession
			p.scoreLabel.SetText(fmt.Sprintf(
				"Score: %d / %d  (%.0f%% accuracy)",
				s.Score, s.Total, s.Accuracy(),
			))
		}

		// Persist to disk and, for wrong answers, add to the vocab list so
		// the user can review it later.
		if p.main.store != nil && p.main.activeProfile != nil {
			_ = p.main.store.RecordAttempt(p.main.activeProfile.Name, result)

			if !isCorrect && correctTranslation != "" {
				entry := models.VocabEntry{
					English:       p.currentPrompt,
					WrongAnswer:   answer,
					CorrectAnswer: correctTranslation,
					Language:      p.language.Name,
					AddedAt:       time.Now(),
				}
				_ = p.main.store.AddVocabEntry(p.main.activeProfile.Name, entry)
			}
		}

		p.nextBtn.Show()
		p.main.setStatus("See feedback below, then press Next Question.")
	}()
}

// setLoading toggles the infinite progress bar and the submit button.
// Always called from the main goroutine (or right before/after a go block)
// so Fyne widget updates stay on the correct thread.
func (p *PracticeScreen) setLoading(on bool) {
	if on {
		p.loading.Show()
		p.loading.Start()
		p.submitBtn.Disable() // prevent double-submit while waiting for the API
	} else {
		p.loading.Stop()
		p.loading.Hide()
		p.submitBtn.Enable()
	}
}
