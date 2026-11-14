package workflowhttp

import (
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	du "erpbadger/internal/domain/workflow"
	uc "erpbadger/internal/usecase/workflow"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers {
	return &Handlers{svc: svc}
}
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
func (h *Handlers) UpdateWorkflowSteps(c echo.Context) error {
	id := atoi(c.Param("id"))

	var req struct {
		Steps []du.WorkflowStep `json:"steps"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(400, echo.Map{"error": "bad payload"})
	}

	if err := h.svc.UpdateWorkflowSteps(
		c.Request().Context(),
		id,
		req.Steps,
	); err != nil {
		return c.JSON(400, echo.Map{"error": err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handlers) GetWorkflowByID(c echo.Context) error {
	wf, err := h.svc.GetWorkflowByID(
		c.Request().Context(),
		atoi(c.Param("id")),
	)
	if err != nil {
		return c.JSON(404, echo.Map{"error": err.Error()})
	}
	return c.JSON(200, echo.Map{"data": wf})
}
func (h *Handlers) Workflows(c echo.Context) error {
	data, err := h.svc.GetWorkflows(c.Request().Context())
	if err != nil {
		return c.JSON(500, echo.Map{"error": err.Error()})
	}
	return c.JSON(200, echo.Map{"data": data})
}

func (h *Handlers) GetWorkflowForResource(c echo.Context) error {
	wf, err := h.svc.GetWorkflowForResource(
		c.Request().Context(),
		c.Param("resource"),
	)
	if err != nil {
		return c.JSON(400, echo.Map{"error": err.Error()})
	}
	return c.JSON(200, echo.Map{"data": wf})
}

func (h *Handlers) CreateWorkflow(c echo.Context) error {
	var req struct {
		Name     string `json:"name"`
		Resource string `json:"resource"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(400, echo.Map{"error": "bad payload"})
	}

	if err := h.svc.CreateWorkflow(
		c.Request().Context(),
		req.Name,
		req.Resource,
	); err != nil {
		return c.JSON(400, echo.Map{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

