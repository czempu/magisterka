package organizationhttp

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	uc "erpbadger/internal/usecase/organization"
)

type Handlers struct {
	svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }









func (h *Handlers) DeleteOrg(c echo.Context) error {
	id := c.Param("orgid")
	idint, errai := strconv.Atoi(id)

	if errai != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Niepoprawny ID kontraktu zatrudnienia"})
	}


	if err := h.svc.DeleteOrg(c.Request().Context(), idint); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Usunięto Jednostkę"})
}











func (h *Handlers) OrgUnits(c echo.Context) error {
	orgs, err := h.svc.GetOrgUnits(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": orgs})
}

func (h *Handlers) AddOrgUnit(c echo.Context) error {
	var req struct {
		Name     string `json:"name"`
		Code     string `json:"code"`
		ParentID *int   `json:"parent_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}

	info, err := h.svc.AddOrgUnit(c.Request().Context(), req.Name, req.Code, req.ParentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) AssignUser(c echo.Context) error {
	var req struct {
		UserID    int    `json:"user_id"`
		OrgUnitID int    `json:"org_unit_id"`
		Title     string `json:"title"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}
	req.IsPrimary = true
	info, err := h.svc.AssignUser(c.Request().Context(), req.UserID, req.OrgUnitID, req.Title, req.IsPrimary)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) SetManager(c echo.Context) error {
	var req struct {
		OrgUnitID int  `json:"org_unit_id"`
		UserID    int  `json:"user_id"`
		IsDefault bool `json:"is_default"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
	}

	info, err := h.svc.SetManager(c.Request().Context(), req.OrgUnitID, req.UserID, req.IsDefault)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) GetManagerForUser(c echo.Context) error {
	idParam := c.Param("userid")
	userID, _ := strconv.Atoi(idParam)
	managerID, err := h.svc.GetManagerForUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"manager_id": managerID})
}
func (h *Handlers) UpdateOrgUnit(c echo.Context) error {
    var req struct {
        ID       int    `json:"id"`
        Name     string `json:"name"`
        Code     string `json:"code"`
        ParentID *int   `json:"parent_id"`
    }
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    info, err := h.svc.UpdateOrgUnit(c.Request().Context(), req.ID, req.Name, req.Code, req.ParentID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": info})
}

func (h *Handlers) GetManagerForUnit(c echo.Context) error {
    idParam := c.Param("orgunitid")
    orgID, _ := strconv.Atoi(idParam)

    managerID, err := h.svc.GetManagerForUnit(c.Request().Context(), orgID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"manager_id": managerID})
}
func (h *Handlers) ListUserMemberships(c echo.Context) error {
    mems, err := h.svc.ListUserMemberships(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mems})
}

