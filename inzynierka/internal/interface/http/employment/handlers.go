package employmenthttp

import (
	"net/http"
	"strconv"
	//"time"

	"github.com/labstack/echo/v4"
	uc "erpbadger/internal/usecase/employment"
	du "erpbadger/internal/domain/employment"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }

func getUserID(c echo.Context) int {
	id, _ := strconv.Atoi(c.Param("id"))
	return id
}

func (h *Handlers) ListTerms(c echo.Context) error {
	data, err := h.svc.GetTerms(c.Request().Context(), getUserID(c))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

func (h *Handlers) AddTerm(c echo.Context) error {
	var req du.EmploymentTermAdd
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Błędne dane"})
	}
	
	if len(req.Position) > 30 {
			return c.JSON(http.StatusBadRequest, echo.Map{"data": "Nazwa stanowiska przekracza dozwoloną długość (30)"})
	}

	//req.EffectiveFrom = time.Now().String()
	req.UserID = getUserID(c)
	if err := h.svc.AddTerm(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": "Dodano nowy wymiar czasu pracy"})
}
func (h *Handlers) UpdateTerm(c echo.Context) error {
	var req du.EmploymentTerm
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Błędne dane"})
	}


	if len(req.Position) > 30 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Nazwa stanowiska przekracza dozwoloną długość (30)"})
	}

	termID, err := strconv.Atoi(c.Param("term_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Niepoprawny ID okresu"})
	}

	req.ID = termID

	if err := h.svc.UpdateTerm(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": "Zaktualizowano okres zatrudnienia"})
}
func (h *Handlers) DeleteTerm(c echo.Context) error {
	id := c.Param("term_id")
	idint, errai := strconv.Atoi(id)

	if errai != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Niepoprawny ID kontraktu zatrudnienia"})
	}


	if err := h.svc.DeleteTerm(c.Request().Context(), idint); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Usunięto kontrakt"})
}
