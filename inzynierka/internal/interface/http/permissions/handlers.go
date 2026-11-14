package permissionshttp

import (
    //"net/http"
    "strconv"

    "github.com/labstack/echo/v4"

    du "erpbadger/internal/domain/permissions"
    uc "erpbadger/internal/usecase/permissions"
)

type Handlers struct {
    svc uc.Usecase
}

func New(svc uc.Usecase) *Handlers {
    return &Handlers{svc: svc}
}

func parseID(c echo.Context, name string) (int, error) {
    v := c.Param(name)
    return strconv.Atoi(v)
}

//
// === STARE ENDPOINTY — pozostawiam jeśli chcesz je zachować ===
//

func (h *Handlers) SystemRoles(c echo.Context) error {
    out, err := h.svc.GetSystemRoles(c.Request().Context())
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, echo.Map{"data": out})
}

func (h *Handlers) UserRole(c echo.Context) error {
    id, _ := parseID(c, "id")
    out, err := h.svc.GetRoleByUserID(c.Request().Context(), id)
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, echo.Map{"data": out})
}

func (h *Handlers) UserRoleChange(c echo.Context) error {
    id, _ := parseID(c, "id")
    var body struct{ Role int `json:"roleid"` }

    if err := c.Bind(&body); err != nil {
        return c.JSON(400, "bad request")
    }
    msg, err := h.svc.UserRoleChange(c.Request().Context(), id, body.Role)
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, echo.Map{"data": msg})
}

//
// === NOWE REST: ROLES ===
//
func (h *Handlers) RolePermissionsList(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("roleId"))

    ids, err := h.svc.GetPermissionsAssignedToRole(c.Request().Context(), id)
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, ids)
}


func (h *Handlers) RolesList(c echo.Context) error {
    out, err := h.svc.GetRoles(c.Request().Context())
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, out)
}

func (h *Handlers) RoleGet(c echo.Context) error {
    id, _ := parseID(c, "id")
    out, err := h.svc.GetRoleByID(c.Request().Context(), id)
    if err != nil {
        return c.JSON(500, err.Error())
    }
    if out == nil {
        return c.JSON(404, "not found")
    }
    return c.JSON(200, out)
}

func (h *Handlers) RoleCreate(c echo.Context) error {
    var body struct {
        Slug string `json:"slug"`
        Name string `json:"name"`
    }
    if err := c.Bind(&body); err != nil {
        return c.JSON(400, "bad input")
    }

    id, err := h.svc.CreateRole(c.Request().Context(), body.Slug, body.Name)
    if err != nil {
        return c.JSON(400, err.Error())
    }
    return c.JSON(201, echo.Map{"id": id})
}

func (h *Handlers) RoleUpdate(c echo.Context) error {
    id, _ := parseID(c, "id")
    var body struct {
        Slug string `json:"slug"`
        Name string `json:"name"`
    }
    if err := c.Bind(&body); err != nil {
        return c.JSON(400, "bad input")
    }

    if err := h.svc.UpdateRole(c.Request().Context(), id, body.Slug, body.Name); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

func (h *Handlers) RoleDelete(c echo.Context) error {
    id, _ := parseID(c, "id")
    if err := h.svc.DeleteRole(c.Request().Context(), id); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

//
// === NOWE REST: PERMISSIONS CRUD ===
//

func (h *Handlers) PermissionsList(c echo.Context) error {
    out, err := h.svc.GetPermissions(c.Request().Context())
    if err != nil {
        return c.JSON(500, err.Error())
    }
    return c.JSON(200, out)
}

func (h *Handlers) PermissionGet(c echo.Context) error {
    id, _ := parseID(c, "id")
    out, err := h.svc.GetPermissionByID(c.Request().Context(), id)
    if err != nil {
        return c.JSON(500, err.Error())
    }
    if out == nil {
        return c.JSON(404, "not found")
    }
    return c.JSON(200, out)
}

func (h *Handlers) PermissionCreate(c echo.Context) error {
    var in du.Permission
    if err := c.Bind(&in); err != nil {
        return c.JSON(400, "bad input")
    }

    id, err := h.svc.CreatePermission(c.Request().Context(), in)
    if err != nil {
        return c.JSON(400, err.Error())
    }
    return c.JSON(201, echo.Map{"id": id})
}

func (h *Handlers) PermissionUpdate(c echo.Context) error {
    id, _ := parseID(c, "id")
    var in du.Permission

    if err := c.Bind(&in); err != nil {
        return c.JSON(400, "bad input")
    }

    in.ID = id
    if err := h.svc.UpdatePermission(c.Request().Context(), in); err != nil {
        return c.JSON(400, err.Error())
    }

    return c.NoContent(200)
}

func (h *Handlers) PermissionDelete(c echo.Context) error {
    id, _ := parseID(c, "id")
    if err := h.svc.DeletePermission(c.Request().Context(), id); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

//
// === NOWE REST: RELATIONS ===
//

func (h *Handlers) RoleAssignPermission(c echo.Context) error {
    roleID, _ := parseID(c, "roleId")
    permID, _ := parseID(c, "permId")

    if err := h.svc.AssignPermissionToRole(c.Request().Context(), roleID, permID); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

func (h *Handlers) RoleRemovePermission(c echo.Context) error {
    roleID, _ := parseID(c, "roleId")
    permID, _ := parseID(c, "permId")

    if err := h.svc.RemovePermissionFromRole(c.Request().Context(), roleID, permID); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

func (h *Handlers) UserAssignRole(c echo.Context) error {
    userID, _ := parseID(c, "userId")
    roleID, _ := parseID(c, "roleId")

    if err := h.svc.AssignRoleToUser(c.Request().Context(), userID, roleID); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

func (h *Handlers) UserRemoveRole(c echo.Context) error {
    userID, _ := parseID(c, "userId")
    roleID, _ := parseID(c, "roleId")

    if err := h.svc.RemoveRoleFromUser(c.Request().Context(), userID, roleID); err != nil {
        return c.JSON(400, err.Error())
    }
    return c.NoContent(200)
}

