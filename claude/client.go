// Package claude wraps the Google Gemini HTTP API so the rest of the app
// doesn't have to think about JSON marshalling, headers, or error shapes.
//
// The package is still named "claude" so every other file that imports it
// doesn't need updating — only the underlying transport changed to Gemini.
//
// Reference: https://ai.google.dev/api/generate-content
package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ModelName is the Gemini model variant we target. As of early 2026 this is
// available on the free tier, which is all we need for a student project.
const ModelName = "gemini-2.5-flash"

// MaxTokens caps how long each AI response can be. 1 024 tokens is plenty
// for a single translated sentence plus a short feedback paragraph.
const MaxTokens = 1024

// apiEndpoint builds the full REST URL for generateContent. Gemini authenticates
// via a query-string key rather than an Authorization header, so we embed
// the key directly into the URL here.
func apiEndpoint(apiKey string) string {
	return fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		ModelName, apiKey,
	)
}

// Client holds the API key and a reusable HTTP client. Create one with
// NewClient and then call Send as many times as you need.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// ─── Request types ────────────────────────────────────────────────────────────

// Part is the smallest unit of content in a Gemini message — just a text blob.
type Part struct {
	Text string `json:"text"`
}

// GeminiContent represents one turn in the conversation (either "user" or "model").
// Gemini uses an array of Parts so it could technically support multimodal input,
// but we always send exactly one text Part.
type GeminiContent struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// SystemInstruction carries the "system prompt" — background context the model
// reads before the user's message. This is separate from the conversation history
// so Gemini can treat it differently from ordinary turns.
type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

// GenerationConfig lets us tweak sampling behaviour. For now we only set
// the output-token ceiling; temperature etc. are left at Gemini's defaults.
type GenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

// RequestBody is the top-level JSON object we POST to the Gemini endpoint.
type RequestBody struct {
	SystemInstruction SystemInstruction `json:"system_instruction"`
	Contents          []GeminiContent   `json:"contents"`
	GenerationConfig  GenerationConfig  `json:"generationConfig"`
}

// ─── Response types ───────────────────────────────────────────────────────────

// ResponsePart mirrors Part but lives on the response side of the wire.
type ResponsePart struct {
	Text string `json:"text"`
}

// ResponseContent groups the Parts that make up one candidate reply.
type ResponseContent struct {
	Parts []ResponsePart `json:"parts"`
}

// Candidate is a single completion the model produced. Gemini can return
// multiple candidates but we always just use index 0.
type Candidate struct {
	Content ResponseContent `json:"content"`
}

// GeminiError is the structured error block Gemini returns when something
// goes wrong on the API side (bad key, quota exceeded, etc.).
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ResponseBody is the top-level JSON object we get back from Gemini.
// Either Candidates will be populated (success) or Error will be non-nil (failure).
type ResponseBody struct {
	Candidates []Candidate  `json:"candidates"`
	Error      *GeminiError `json:"error,omitempty"`
}

// NewClient creates a ready-to-use Gemini client. It does a quick sanity-check
// on the key format but does NOT make a network call — the first real call
// happens when you call Send.
//
// Returns an error if apiKey is blank.
func NewClient(apiKey string) (*Client, error) {
	// Refuse to build a client we already know will fail every request.
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	return &Client{
		apiKey: apiKey,
		// 30 seconds is generous for a single completion request, but avoids
		// hanging the UI forever if the user's network is flaky.
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Send fires a single completion request at Gemini and returns the text of
// the first candidate. systemPrompt sets the model's persona/task; userMessage
// is the thing we actually want it to respond to.
//
// Common failure modes:
//   - empty userMessage  → validation error, no network call made
//   - bad API key        → API error with code 400 or 403
//   - network timeout    → wrapped net/http error
func (c *Client) Send(systemPrompt string, userMessage string) (string, error) {
	// Don't bother hitting the network for an empty message.
	if strings.TrimSpace(userMessage) == "" {
		return "", fmt.Errorf("user message cannot be empty")
	}

	// Assemble the request payload in Gemini's expected shape.
	reqBody := RequestBody{
		SystemInstruction: SystemInstruction{
			Parts: []Part{{Text: systemPrompt}},
		},
		Contents: []GeminiContent{
			{
				// "user" is the only role we send; the model's prior reply
				// isn't passed back because each exercise is stateless.
				Role:  "user",
				Parts: []Part{{Text: userMessage}},
			},
		},
		GenerationConfig: GenerationConfig{
			MaxOutputTokens: MaxTokens,
		},
	}

	// Marshal to JSON — this really shouldn't fail given the fixed struct shape,
	// but we handle it anyway to be safe.
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	// Build the POST request. The API key rides in the URL (Gemini style),
	// not in a Bearer token header.
	req, err := http.NewRequest("POST", apiEndpoint(c.apiKey), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Actually send it.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Surface a user-friendly hint alongside the raw error.
		return "", fmt.Errorf("request failed — check your internet connection: %w", err)
	}
	defer resp.Body.Close()

	// Slurp the whole body before parsing so we have it for debugging if needed.
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal into our response struct.
	var respBody ResponseBody
	if err := json.Unmarshal(respBytes, &respBody); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// API-level errors (invalid key, quota, etc.) come back as HTTP 200 with
	// a populated Error field, so we check that before the status code.
	if respBody.Error != nil {
		return "", fmt.Errorf("API error %d: %s", respBody.Error.Code, respBody.Error.Message)
	}

	// Non-200 that didn't include a structured error — catch it as a fallback.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Gemini can return zero candidates if the response was filtered for safety.
	if len(respBody.Candidates) == 0 || len(respBody.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("API returned empty response")
	}

	// Return the text from the first (and usually only) candidate.
	return respBody.Candidates[0].Content.Parts[0].Text, nil
}
