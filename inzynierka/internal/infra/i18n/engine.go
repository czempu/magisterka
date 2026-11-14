package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	domain "erpbadger/internal/domain/i18n"
	"erpbadger/locales"
)

type Engine struct {
	mu           sync.RWMutex
	defaultLang  string
	languages    []domain.Language
	langMap      map[string]domain.Language
	dictionaries map[string]map[string]string
}

var knownMeta = map[string]domain.Language{
	"pl": {Code: "pl", Name: "Polski", Flag: "🇵🇱"},
	"en": {Code: "en", Name: "English", Flag: "🇬🇧"},
	"de": {Code: "de", Name: "Deutsch", Flag: "🇩🇪"},
	"es": {Code: "es", Name: "Español", Flag: "🇪🇸"},
	"fr": {Code: "fr", Name: "Français", Flag: "🇫🇷"},
	"uk": {Code: "uk", Name: "Українська", Flag: "🇺🇦"},
	"it": {Code: "it", Name: "Italiano", Flag: "🇮🇹"},
	"cs": {Code: "cs", Name: "Čeština", Flag: "🇨🇿"},
}

// NewEngine initializes the translation engine by loading all embedded JSON files.
func NewEngine() (domain.Translator, error) {
	e := &Engine{
		defaultLang:  "pl",
		langMap:      make(map[string]domain.Language),
		dictionaries: make(map[string]map[string]string),
	}

	// Discover languages from directory names in locales.FS
	entries, err := fs.ReadDir(locales.FS, ".")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				code := strings.ToLower(entry.Name())
				meta, ok := knownMeta[code]
				if !ok {
					meta = domain.Language{
						Code: code,
						Name: strings.ToUpper(code),
						Flag: "🌐",
					}
				}
				e.languages = append(e.languages, meta)
				e.langMap[code] = meta
				e.dictionaries[code] = make(map[string]string)
			}
		}
	}

	// Fallback if no dirs found
	if len(e.languages) == 0 {
		for _, code := range []string{"pl", "en", "de"} {
			meta := knownMeta[code]
			e.languages = append(e.languages, meta)
			e.langMap[code] = meta
			e.dictionaries[code] = make(map[string]string)
		}
	}

	if err := e.loadTranslations(); err != nil {
		return nil, fmt.Errorf("błąd ładowania tłumaczeń: %w", err)
	}

	return e, nil
}

func (e *Engine) loadTranslations() error {

	err := fs.WalkDir(locales.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return nil
		}

		lang := parts[0]
		if _, exists := e.langMap[lang]; !exists {
			return nil
		}

		fileName := parts[1]
		moduleName := strings.TrimSuffix(fileName, ".json")

		data, err := fs.ReadFile(locales.FS, path)
		if err != nil {
			return fmt.Errorf("błąd czytania pliku %s: %w", path, err)
		}

		var rawMap map[string]any
		if err := json.Unmarshal(data, &rawMap); err != nil {
			return fmt.Errorf("błąd parsowania JSON w %s: %w", path, err)
		}

		e.flattenMap(lang, moduleName, rawMap)
		return nil
	})

	return err
}

func (e *Engine) flattenMap(lang, prefix string, m map[string]any) {
	for k, v := range m {
		fullKey := prefix + "." + k
		switch val := v.(type) {
		case string:
			e.dictionaries[lang][fullKey] = val
			// Also store unprefixed if no collision to allow short keys
			if _, exists := e.dictionaries[lang][k]; !exists {
				e.dictionaries[lang][k] = val
			}
		case map[string]any:
			e.flattenMap(lang, fullKey, val)
		}
	}
}

func (e *Engine) T(lang, key string, args ...any) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	lang = strings.ToLower(strings.TrimSpace(lang))
	if !e.IsSupported(lang) {
		lang = e.defaultLang
	}

	dict, ok := e.dictionaries[lang]
	var text string
	if ok {
		text, ok = dict[key]
	}

	if !ok || text == "" {
		// Fallback to default language
		if fallbackDict, hasDef := e.dictionaries[e.defaultLang]; hasDef {
			text = fallbackDict[key]
		}
	}

	if text == "" {
		// Ultimate fallback: return key
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}

	return text
}

func (e *Engine) SupportedLanguages() []domain.Language {
	return e.languages
}

func (e *Engine) IsSupported(lang string) bool {
	_, ok := e.langMap[strings.ToLower(strings.TrimSpace(lang))]
	return ok
}

func (e *Engine) DefaultLanguage() string {
	return e.defaultLang
}

func (e *Engine) GetDictionary(lang string) map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	lang = strings.ToLower(strings.TrimSpace(lang))
	if !e.IsSupported(lang) {
		lang = e.defaultLang
	}

	result := make(map[string]string, len(e.dictionaries[lang]))
	for k, v := range e.dictionaries[lang] {
		result[k] = v
	}
	return result
}
