package models

// SingleTranslationRequest represents a request to translate a single text message.
type SingleTranslationRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Text   string `json:"text"`
}

// SingleTranslationResponse represents the response containing the translated text.
type SingleTranslationResponse struct {
	TranslatedText string `json:"translated_text"`
}

// LocaleItem represents a localized key-value entry.
type LocaleItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// LocaleFileRequest represents a batch request to translate multiple localization entries.
type LocaleFileRequest struct {
	Source string       `json:"source"`
	Target string       `json:"target"`
	Texts  []LocaleItem `json:"texts"`
}

// LocaleFileResponse represents the response containing translated localization entries.
type LocaleFileResponse struct {
	Texts []LocaleItem `json:"texts"`
}

// HealthResponse represents the health check status payload.
type HealthResponse struct {
	Status string `json:"status"`
	Model  string `json:"model"`
}
