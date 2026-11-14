package approvalhttp

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	uc "erpbadger/internal/usecase/approval"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }

func (h *Handlers) Register(g *echo.Group) {

}

func (h *Handlers) Create(c echo.Context) error {
	var req struct {
		Resource   string `json:"resource"`
		TargetID   int64  `json:"target_id"`
		Requester  int    `json:"requester_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}

	info, err := h.svc.CreateApproval(c.Request().Context(), req.Resource, req.TargetID, req.Requester)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) Decision(c echo.Context) error {
	stepID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		ApproverID int     `json:"approver_id"`
		Decision   string  `json:"decision"`
		Note       *string `json:"note"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}
	info, err := h.svc.Decision(c.Request().Context(), stepID, body.ApproverID, body.Decision, body.Note)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) Pending(c echo.Context) error {
	approver, _ := strconv.Atoi(c.Param("approver"))
	data, err := h.svc.ListPending(c.Request().Context(), approver)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}
