

package moduleshttp

import (
    "net/http"
	"strconv"	
    "fmt"

    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
    uc "erpbadger/internal/usecase/modules"
    "erpbadger/internal/interface/http/middleware"
	"erpbadger/config"
)

func getUser(c echo.Context) int {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }
    return id 

}

var p = fmt.Println

var jwtSecret = []byte(config.Secret)

type Handlers struct {
    svc uc.Usecase
}
func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }





//modules
func (h *Handlers) SystemModules(c echo.Context) error { 
	 //sd := c.Get("session").(*middleware.SessionData)

    mods, err := h.svc.GetSystemModules(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}


//modules
func (h *Handlers) UserModules(c echo.Context) error {
    userToken := c.Get("user").(*jwt.Token)
    claims := userToken.Claims.(jwt.MapClaims)
    userID := int(claims["id"].(float64))

    mods, err := h.svc.GetUserModules(c.Request().Context(), userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}










// Module DTO — pojedynczy moduł użytkownika.
type Module struct {
	ID          int64  `json:"id"           example:"1"`
	ModuleName  string `json:"module_name"  example:"users"`
	Description string `json:"description"  example:"użytkownicy"`
	Role        string `json:"role"         example:"test"`
}

// Module DTO
// ModulesResponse — odpowiedź z listą modułów.
type ModulesResponse struct {
	Data []Module `json:"data"`
}

// HANDLER + adnotacje Swaggera

// GetMyModules godoc
// @Summary      Lista modułów użytkownika
// @Description  Zwraca moduły przypisane do aktualnie zalogowanego użytkownika.
// @Tags         Me
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ModulesResponse
// @Router       /api/me/modules [get]
func (h *Handlers) MeModules(c echo.Context) error { //modules
sd := c.Get("session").(*middleware.SessionData)
fmt.Println("handler - MeModules")
    mods, err := h.svc.GetMeModules(c.Request().Context(), sd.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}


