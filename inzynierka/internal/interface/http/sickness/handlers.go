package sicknesshttp

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	uc "erpbadger/internal/usecase/sickness"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }

// ------- PAYLOADY ---------

type SicknessCreatePayload struct {
	FromDate          string `json:"from_date"`             // YYYY-MM-DD
	ToDate            string `json:"to_date"`               // YYYY-MM-DD
	LegalBasis        string `json:"legal_basis"`           // required
	CertificateNumber string `json:"certificate_number"`    // optional
	PercentCoverage   int    `json:"percent_coverage"`      // 0 = auto (80/100)
	IsAccident        bool   `json:"is_occupational_accident"`
}

// ------- HANDLERY API -------

// GET /api/sickness?employee_id=&status=
func (h *Handlers) List(c echo.Context) error {
	var employeeID *int
	var status *string

	if eidStr := c.QueryParam("employee_id"); eidStr != "" {
		if eid, err := strconv.Atoi(eidStr); err == nil && eid > 0 {
			employeeID = &eid
		}
	}
	if st := c.QueryParam("status"); st != "" {
		status = &st
	}

	res, err := h.svc.List(c.Request().Context(), employeeID, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": res})
}

// GET /api/sickness/:id
func (h *Handlers) Get(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id"})
	}

	s, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if s == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "nie znaleziono chorobowego"})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": s})
}

// POST /api/sickness
// tworzy zgłoszenie chorobowe + odpala workflow "l4"
func (h *Handlers) Create(c echo.Context) error {
	idstring := c.Param("id")
	id,atierr := strconv.Atoi(idstring)
	if atierr != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id użytkonika"})
	}
	var payload SicknessCreatePayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}

	resp, err := h.svc.Create(
		c.Request().Context(),
		id,
		payload.FromDate,
		payload.ToDate,
		payload.LegalBasis,
		payload.CertificateNumber,
		payload.PercentCoverage,
		payload.IsAccident,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": resp})
}

// DELETE /api/sickness/:id
func (h *Handlers) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id"})
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": "usunięto zgłoszenie chorobowe"})
}

// PUT /api/sickness/:id/status
func (h *Handlers) UpdateStatus(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}

	if err := h.svc.UpdateStatus(c.Request().Context(), id, req.Status); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": "zaktualizowano status chorobowego"})
}

