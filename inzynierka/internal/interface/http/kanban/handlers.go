package kanban

import (
	"net/http"
	"strconv"

	du "erpbadger/internal/domain/kanban"
	uc "erpbadger/internal/usecase/kanban"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) GetTickets(c echo.Context) error {
	tickets, err := h.svc.GetTickets(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tickets)
}

func (h *Handlers) CreateTicket(c echo.Context) error {
	var t du.Ticket
	if err := c.Bind(&t); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe dane"})
	}
	if err := h.svc.CreateTicket(c.Request().Context(), &t); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *Handlers) UpdateTicket(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe ID zadania"})
	}
	var t du.Ticket
	if err := c.Bind(&t); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe dane"})
	}
	if err := h.svc.UpdateTicket(c.Request().Context(), id, &t); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handlers) DeleteTicket(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe ID zadania"})
	}
	if err := h.svc.DeleteTicket(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handlers) GetCalendar(c echo.Context) error {
	entries, err := h.svc.GetCalendar(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, entries)
}

func (h *Handlers) UpdateCalendar(c echo.Context) error {
	var entry du.CalendarEntry
	if err := c.Bind(&entry); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe dane"})
	}
	if err := h.svc.UpdateCalendar(c.Request().Context(), &entry); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, entry)
}
