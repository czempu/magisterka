
package rcphttp

import (
    "net/http"
	"strconv"	
    "fmt"
	"time"
    "github.com/labstack/echo/v4"
    uc "erpbadger/internal/usecase/rcp"
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





//RCP
//WORK START
func (h *Handlers) MeWorkStart(c echo.Context) error {
sd := c.Get("session").(*middleware.SessionData)
    mods, err := h.svc.PostMeWorkStart(c.Request().Context(), sd.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}
//work status


//RCP
func (h *Handlers) MeWorkStatus(c echo.Context) error {
sd := c.Get("session").(*middleware.SessionData)
fmt.Println("handler - MeModules")
    mods, err := h.svc.GetMeWorkStatus(c.Request().Context(), sd.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}
//work stop

//RCP
func (h *Handlers) MeWorkStop(c echo.Context) error {
sd := c.Get("session").(*middleware.SessionData)
fmt.Println("handler - MeModules")
    mods, err := h.svc.PostMeWorkStop(c.Request().Context(), sd.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}


//RCP
func (h *Handlers) MeWorkTimesheet(c echo.Context) error {
sd := c.Get("session").(*middleware.SessionData)
fmt.Println("handler - MeModules")
    mods, err := h.svc.GetMeWorkTimesheet(c.Request().Context(), sd.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}



// GET /internal/cron/rcp/daily-totals?day=2025-11-12
// - bez parametru day -> liczy wczorajszy dzień
// - z parametrem day=YYYY-MM-DD -> liczy dokładnie tę datę
func (h *Handlers) RecalculateDailyTotals(c echo.Context) error {
	dayStr := c.QueryParam("day")

	var (
		day time.Time
		err error
	)

	if dayStr == "" {
		now := time.Now().In(time.Local)
		day = now.AddDate(0, 0, -1)
	} else {
		day, err = time.Parse("2006-01-02", dayStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "nieprawidłowy format daty, użyj YYYY-MM-DD",
			})
		}
	}

	if err := h.svc.RecalculateDailyTotalsForDate(c.Request().Context(), day); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "ok",
		"day":    day.Format("2006-01-02"),
	})
}
