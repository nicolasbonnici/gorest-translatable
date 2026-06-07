package translatable

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mockTranslator struct {
	result *TranslationResult
	err    error
}

func (m *mockTranslator) Translate(_ context.Context, _, _ string, _ *uuid.UUID) (*TranslationResult, error) {
	return m.result, m.err
}

func newTranslateApp(t Translator) *fiber.App {
	app := fiber.New()
	app.Post("/translations/:type/:id/translate", (&TranslatableResource{translator: &t}).Translate)
	return app
}

func newTranslateAppNil() *fiber.App {
	app := fiber.New()
	app.Post("/translations/:type/:id/translate", (&TranslatableResource{translator: nil}).Translate)
	return app
}

func TestTranslate_NotConfigured(t *testing.T) {
	app := newTranslateAppNil()

	req := httptest.NewRequest("POST", "/translations/post/abc/translate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}

func TestTranslate_TranslatorError(t *testing.T) {
	app := newTranslateApp(&mockTranslator{err: errors.New("ai unavailable")})

	req := httptest.NewRequest("POST", "/translations/post/abc/translate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

func TestTranslate_Success(t *testing.T) {
	expected := &TranslationResult{
		Translated: []string{"fr", "es"},
		Skipped:    []string{"de"},
		Failed:     []string{},
	}
	app := newTranslateApp(&mockTranslator{result: expected})

	req := httptest.NewRequest("POST", "/translations/post/"+uuid.New().String()+"/translate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var got TranslationResult
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(got.Translated) != 2 || got.Translated[0] != "fr" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
