package i18nhttp

import (
	"net/http"
	"strings"

	domain "erpbadger/internal/domain/i18n"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	translator domain.Translator
}

func New(translator domain.Translator) *Handlers {
	return &Handlers{translator: translator}
}

// GetTranslationsHandler returns the full dictionary for the given language.
//
// @Summary Pobierz słownik tłumaczeń
// @Description Zwraca płaski słownik klucz-wartość dla wybranego języka
// @Tags i18n
// @Param lang path string true "Kod języka (pl, en, de)"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/i18n/{lang} [get]
func (h *Handlers) GetTranslations(c echo.Context) error {
	lang := c.Param("lang")
	if !h.translator.IsSupported(lang) {
		lang = h.translator.DefaultLanguage()
	}
	return c.JSON(http.StatusOK, h.translator.GetDictionary(lang))
}

// GetLanguagesHandler returns all supported languages.
//
// @Summary Pobierz listę wspieranych języków
// @Description Zwraca listę języków z kodem, nazwą i flagą
// @Tags i18n
// @Produce json
// @Success 200 {array} domain.Language
// @Router /api/i18n/languages [get]
func (h *Handlers) GetLanguages(c echo.Context) error {
	return c.JSON(http.StatusOK, h.translator.SupportedLanguages())
}

// Middleware resolves the preferred language from query, cookie, or Accept-Language header.
func Middleware(translator domain.Translator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			lang := ""

			// 1. Query parameter: ?lang=
			if q := c.QueryParam("lang"); q != "" {
				lang = q
			}

			// 2. Cookie: lang=
			if lang == "" {
				if cookie, err := c.Cookie("lang"); err == nil && cookie.Value != "" {
					lang = cookie.Value
				}
			}

			// 3. Header: Accept-Language
			if lang == "" {
				accept := c.Request().Header.Get("Accept-Language")
				if accept != "" {
					parts := strings.Split(accept, ",")
					if len(parts) > 0 {
						primary := strings.TrimSpace(parts[0])
						if len(primary) >= 2 {
							lang = primary[:2]
						}
					}
				}
			}

			lang = strings.ToLower(lang)
			if !translator.IsSupported(lang) {
				lang = translator.DefaultLanguage()
			}

			// Store in Echo context
			c.Set("lang", lang)

			return next(c)
		}
	}
}
