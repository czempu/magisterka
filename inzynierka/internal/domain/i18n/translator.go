package i18n

// Language represents a supported system language with metadata.
type Language struct {
	Code string `json:"code"` // e.g. "pl", "en", "de"
	Name string `json:"name"` // e.g. "Polski", "English", "Deutsch"
	Flag string `json:"flag"` // e.g. "🇵🇱", "🇬🇧", "🇩🇪"
}

// Translator defines the domain port for multilingual translations.
type Translator interface {
	// T translates a key in the given language, optionally formatting with args.
	T(lang, key string, args ...any) string

	// SupportedLanguages returns the list of all configured languages.
	SupportedLanguages() []Language

	// IsSupported checks if a language code is available.
	IsSupported(lang string) bool

	// DefaultLanguage returns the base/default fallback language code (e.g. "pl").
	DefaultLanguage() string

	// GetDictionary returns the flattened map of all keys for a language.
	GetDictionary(lang string) map[string]string
}
