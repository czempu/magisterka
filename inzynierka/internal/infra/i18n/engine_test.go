package i18n_test

import (
	"testing"

	"erpbadger/internal/infra/i18n"
)

func TestEngine_Translations(t *testing.T) {
	engine, err := i18n.NewEngine()
	if err != nil {
		t.Fatalf("failed to initialize i18n engine: %v", err)
	}

	tests := []struct {
		lang string
		key  string
		want string
	}{
		{"pl", "common.save", "Zapisz"},
		{"en", "common.save", "Save"},
		{"de", "common.save", "Speichern"},
		{"pl", "nav.home", "Strona główna"},
		{"en", "nav.home", "Home"},
		{"de", "nav.home", "Startseite"},
		{"pl", "schedule.overtime", "Nadgodziny"},
		{"en", "schedule.overtime", "Overtime"},
		{"de", "schedule.overtime", "Überstunden"},
	}

	for _, tt := range tests {
		got := engine.T(tt.lang, tt.key)
		if got != tt.want {
			t.Errorf("T(%q, %q) = %q, want %q", tt.lang, tt.key, got, tt.want)
		}
	}
}

func TestEngine_FallbackToDefault(t *testing.T) {
	engine, err := i18n.NewEngine()
	if err != nil {
		t.Fatalf("failed to initialize i18n engine: %v", err)
	}

	// Unknown language should fallback to PL
	got := engine.T("fr", "common.save")
	if got != "Zapisz" {
		t.Errorf("expected fallback to PL 'Zapisz', got %q", got)
	}

	// Unknown key should return key itself
	gotUnknown := engine.T("pl", "non_existent_key_123")
	if gotUnknown != "non_existent_key_123" {
		t.Errorf("expected key itself, got %q", gotUnknown)
	}
}

func TestEngine_SupportedLanguages(t *testing.T) {
	engine, err := i18n.NewEngine()
	if err != nil {
		t.Fatalf("failed to initialize i18n engine: %v", err)
	}

	langs := engine.SupportedLanguages()
	if len(langs) != 3 {
		t.Errorf("expected 3 languages, got %d", len(langs))
	}

	if !engine.IsSupported("pl") || !engine.IsSupported("en") || !engine.IsSupported("de") {
		t.Errorf("expected pl, en, de to be supported")
	}

	if engine.IsSupported("es") {
		t.Errorf("es should not be supported yet")
	}
}
