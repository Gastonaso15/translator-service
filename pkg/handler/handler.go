// Package handler implements HTTP handlers for the translation service.
package handler

import (
	"encoding/json"
	"net/http"

	"translator-service/pkg/llm"
	"translator-service/pkg/models"
)

// Handler processes incoming HTTP translation requests.
type Handler struct {
	translator llm.Translator
}

// NewHandler constructs a new Handler with the provided Translator.
func NewHandler(translator llm.Translator) *Handler {
	return &Handler{
		translator: translator,
	}
}

// RegisterRoutes registers all service endpoints onto the given ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/api/v1/translate/message", h.HandleTranslateMessage)
	mux.HandleFunc("/api/v1/translate/localefile", h.HandleTranslateLocaleFile)
}

// HandleHealth reports the operational status and active LLM model.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, models.HealthResponse{
		Status: "ok",
		Model:  h.translator.ModelName(),
	})
}

// HandleTranslateMessage handles translating an individual text message.
func (h *Handler) HandleTranslateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SingleTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text == "" || req.Target == "" {
		http.Error(w, "'text' and 'target' are required fields", http.StatusBadRequest)
		return
	}

	translated, err := h.translator.TranslateSingle(r.Context(), req.Source, req.Target, req.Text)
	if err != nil {
		http.Error(w, "Translation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, models.SingleTranslationResponse{
		TranslatedText: translated,
	})
}

// HandleTranslateLocaleFile translates an entire localization file.
func (h *Handler) HandleTranslateLocaleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LocaleFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Texts) == 0 || req.Target == "" {
		http.Error(w, "'texts' list cannot be empty and 'target' is required", http.StatusBadRequest)
		return
	}

	translatedItems, err := h.translator.TranslateLocaleFile(
		r.Context(),
		req.Source,
		req.Target,
		req.Texts,
	)
	if err != nil {
		http.Error(w, "Translation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, models.LocaleFileResponse{
		Texts: translatedItems,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
