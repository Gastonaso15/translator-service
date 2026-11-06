package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"translator-service/pkg/models"
)

func TestTranslateSingle_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Fatalf("expected path /api/generate, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"response": `{"translation": "Hola Mundo"}`,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-model", 5*time.Second)

	res, err := client.TranslateSingle(context.Background(), "en", "es", "Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hola Mundo" {
		t.Errorf("expected %q, got %q", "Hola Mundo", res)
	}
	if client.ModelName() != "test-model" {
		t.Errorf("expected model %q, got %q", "test-model", client.ModelName())
	}
}

func TestTranslateLocaleFile_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed decoding request body: %v", err)
		}

		prompt, ok := reqBody["prompt"].(string)
		if !ok {
			t.Fatalf("missing prompt")
		}

		// Check if prompt contains item1 or item2 to respond accordingly
		var respItems []models.LocaleItem
		if prompt != "" {
			respItems = append(respItems, models.LocaleItem{ID: "item1", Text: "Item 1 Translated"})
		}

		resp := map[string]interface{}{
			"response": `{"texts": [{"id": "item1", "text": "Item 1 Translated"}]}`,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "test-model", 5*time.Second)

	items := []models.LocaleItem{
		{ID: "item1", Text: "Item 1 Original"},
	}

	translated, err := client.TranslateLocaleFile(context.Background(), "es", "en", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(translated) != 1 || translated[0].Text != "Item 1 Translated" {
		t.Fatalf("unexpected translation results: %+v", translated)
	}
}

func TestTranslateSingle_OllamaError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL, "missing-model", 5*time.Second)

	_, err := client.TranslateSingle(context.Background(), "en", "es", "Hello")
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}
