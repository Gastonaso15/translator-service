package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"translator-service/pkg/models"
)

type mockTranslator struct {
	modelName           string
	translateSingleFunc func(ctx context.Context, source, target, text string) (string, error)
	translateLocaleFunc func(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error)
}

func (m *mockTranslator) ModelName() string {
	return m.modelName
}

func (m *mockTranslator) TranslateSingle(ctx context.Context, source, target, text string) (string, error) {
	if m.translateSingleFunc != nil {
		return m.translateSingleFunc(ctx, source, target, text)
	}
	return "translated: " + text, nil
}

func (m *mockTranslator) TranslateLocaleFile(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error) {
	if m.translateLocaleFunc != nil {
		return m.translateLocaleFunc(ctx, source, target, items)
	}
	res := make([]models.LocaleItem, len(items))
	for i, item := range items {
		res[i] = models.LocaleItem{ID: item.ID, Text: "translated: " + item.Text}
	}
	return res, nil
}

func TestHealthHandler(t *testing.T) {
	mock := &mockTranslator{modelName: "qwen2.5:7b"}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.HandleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp models.HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed decoding health response: %v", err)
	}

	if resp.Status != "ok" || resp.Model != "qwen2.5:7b" {
		t.Errorf("unexpected health response: %+v", resp)
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	mock := &mockTranslator{modelName: "qwen2.5:7b"}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	h.HandleHealth(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rr.Code)
	}
}

func TestTranslateMessageHandler_Success(t *testing.T) {
	mock := &mockTranslator{
		modelName: "test-model",
		translateSingleFunc: func(ctx context.Context, source, target, text string) (string, error) {
			return "Hola Mundo", nil
		},
	}
	h := NewHandler(mock)

	body := `{"source": "en", "target": "es", "text": "Hello World"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/translate/message", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleTranslateMessage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp models.SingleTranslationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}

	if resp.TranslatedText != "Hola Mundo" {
		t.Errorf("expected %q, got %q", "Hola Mundo", resp.TranslatedText)
	}
}

func TestTranslateMessageHandler_Validation(t *testing.T) {
	mock := &mockTranslator{modelName: "test-model"}
	h := NewHandler(mock)

	// Missing target
	body := `{"source": "en", "text": "Hello World"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/translate/message", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleTranslateMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTranslateLocaleFileHandler_Success(t *testing.T) {
	mock := &mockTranslator{
		modelName: "test-model",
		translateLocaleFunc: func(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error) {
			return []models.LocaleItem{
				{ID: "btn.save", Text: "Guardar"},
			}, nil
		},
	}
	h := NewHandler(mock)

	body := `{"source": "en", "target": "es", "texts": [{"id": "btn.save", "text": "Save"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/translate/localefile", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleTranslateLocaleFile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp models.LocaleFileResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}

	if len(resp.Texts) != 1 || resp.Texts[0].Text != "Guardar" {
		t.Errorf("unexpected translated texts: %+v", resp.Texts)
	}
}

func TestTranslateLocaleFileHandler_Error(t *testing.T) {
	mock := &mockTranslator{
		modelName: "test-model",
		translateLocaleFunc: func(ctx context.Context, source, target string, items []models.LocaleItem) ([]models.LocaleItem, error) {
			return nil, errors.New("backend failed")
		},
	}
	h := NewHandler(mock)

	body := `{"source": "en", "target": "es", "texts": [{"id": "btn.save", "text": "Save"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/translate/localefile", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleTranslateLocaleFile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
