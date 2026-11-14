

package leaveshttp

import (
    "net/http"
	"strconv"	
    "fmt"
	//"regexp"
	"time"
    "github.com/labstack/echo/v4"
    uc "erpbadger/internal/usecase/leaves"
  "erpbadger/internal/interface/http/middleware"
  du "erpbadger/internal/domain/leaves"
	"erpbadger/config"

)

func getUser(c echo.Context) int {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }
    return id 

}

type LeaveBalancePayload struct {
    EmployeeID             int    `json:"employee_id"`
    LeaveType              string `json:"leave_type"`
    AccrualDate            string `json:"accrual_date"`
    ExpirationDate         string `json:"expiration_date"`
    InitialBalanceMinutes  int    `json:"initial_balance_minutes"`
    RemainingBalanceMinutes int   `json:"remaining_balance_minutes"`
}

var p = fmt.Println

var jwtSecret = []byte(config.Secret)

type Handlers struct {
    svc uc.Usecase
}
func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }



type RequestAdd struct {     // wnioski 
    Code string `json:"code"`
 StartDate string `json:"startDate"`
       EndDate string `json:"endDate"`
	Unit string `json:"unit"`
	Quantity int `json:"quantity"`
    Note string `json:"note"`
}

// GET /api/leave-balance?employee_id=&leave_type=
func (h *Handlers) LeaveBalanceList(c echo.Context) error {
    var employeeID *int
    var leaveType *string

    if eidStr := c.QueryParam("employee_id"); eidStr != "" {
        if eid, err := strconv.Atoi(eidStr); err == nil && eid > 0 {
            employeeID = &eid
        }
    }
    if lt := c.QueryParam("leave_type"); lt != "" {
        leaveType = &lt
    }

    res, err := h.svc.ListLeaveBalances(c.Request().Context(), employeeID, leaveType)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": res})
}

// POST /api/leave-balance
func (h *Handlers) LeaveBalanceCreate(c echo.Context) error {
    var payload LeaveBalancePayload
    if err := c.Bind(&payload); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    b := du.LeaveBalance{
        EmployeeID:             payload.EmployeeID,
        LeaveType:              payload.LeaveType,
        AccrualDate:            payload.AccrualDate,
        ExpirationDate:         payload.ExpirationDate,
        InitialBalanceMinutes:  payload.InitialBalanceMinutes,
        RemainingBalanceMinutes: payload.RemainingBalanceMinutes,
    }

    id, err := h.svc.CreateLeaveBalance(c.Request().Context(), b)
    if err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": echo.Map{"id": id}})
}

// PUT /api/leave-balance/:id
func (h *Handlers) LeaveBalanceUpdate(c echo.Context) error {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil || id <= 0 {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id"})
    }

    var payload LeaveBalancePayload
    if err := c.Bind(&payload); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }

    b := du.LeaveBalance{
        ID:                      id,
        EmployeeID:              payload.EmployeeID,
        LeaveType:               payload.LeaveType,
        AccrualDate:             payload.AccrualDate,
        ExpirationDate:          payload.ExpirationDate,
        InitialBalanceMinutes:   payload.InitialBalanceMinutes,
        RemainingBalanceMinutes: payload.RemainingBalanceMinutes,
    }

    if err := h.svc.UpdateLeaveBalance(c.Request().Context(), b); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, echo.Map{"data": "zaktualizowano bilans urlopu"})
}

// DELETE /api/leave-balance/:id
func (h *Handlers) LeaveBalanceDelete(c echo.Context) error {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil || id <= 0 {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "nieprawidłowe id"})
    }

    if err := h.svc.DeleteLeaveBalance(c.Request().Context(), id); err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, echo.Map{"data": "usunięto bilans urlopu"})
}

// urlopy
func (h *Handlers) LeaveTypes(c echo.Context) error {
    mods, err := h.svc.GetLeaveTypes(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}	



// urlopy
func (h *Handlers) Request(c echo.Context) error {
	var filter du.UserFilter

	limit, ee := strconv.Atoi(c.QueryParam("limit"))
	if ee != nil {
		limit = 0
	}
	sort := c.QueryParam("sort")

	if c.QueryParam("user_id") != "" {
		fmt.Println("user id przekazano")
		tmp,err := strconv.Atoi(c.QueryParam("user_id"))
			if err == nil {
			filter.ID = &tmp

		}
	} else {
		fmt.Println("userid nie przekazanio")
	}

  sd := c.Get("session").(*middleware.SessionData)
  filter.UserScope.ID = sd.UserID




    mods, err := h.svc.GetRequest(c.Request().Context(), limit, sort, filter)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}


//urlopy
func (h *Handlers) RequestInfo(c echo.Context) error {
	idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
    }

    mods, err := h.svc.GetRequestInfo(c.Request().Context(), id)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": mods})
}

// urlopy
func (h *Handlers) RequestAdd(c echo.Context) error {

sd := c.Get("session").(*middleware.SessionData)

    var req RequestAdd
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "błędne dane"})
    }
    if len(req.Note) > 1000  {
	    return c.JSON(http.StatusInternalServerError, echo.Map{"error": "przekroczono długość notatki (1000 znaków)"})
 }
    if req.Unit == "day" {

		fmt.Println(" leave set to days")
		    
		fmt.Println(" start ")

		fmt.Println(req.StartDate)
		fmt.Println(" end ")

		fmt.Println(req.EndDate)


		// Zakładając, że req to Twoja struktura
layout := "2006-01-02T15:04"

// 1. Walidacja logiczna (czy to w ogóle poprawne daty)
if _, err := time.Parse(layout, req.StartDate); err != nil {
    return fmt.Errorf("nieprawidłowa data rozpoczęcia: %w", err)
}
if _, err := time.Parse(layout, req.EndDate); err != nil {
    return fmt.Errorf("nieprawidłowa data zakończenia: %w", err)
}

// 2. Transformacja stringów "w miejscu"
// Bierzemy YYYY-MM-DD i doklejamy sztywny czas
req.StartDate = req.StartDate[:10] + "T00:00"
req.EndDate   = req.EndDate[:10] + "T23:59"

	}

if req.Unit == "hour" {
layout := "2006-01-02T15:04"

    // 1. Walidacja formatu i poprawności dat
    if _, err := time.Parse(layout, req.StartDate); err != nil {
return c.JSON(http.StatusInternalServerError, echo.Map{"error": "niepoprawny początek urlopu"})
}
    if _, err := time.Parse(layout, req.EndDate); err != nil {
return c.JSON(http.StatusInternalServerError, echo.Map{"error": "niepoprawny koniec urlopu"})
 }

    fmt.Println("Walidacja 'hour' przeszła pomyślnie - ten sam dzień.")

} 



   request, err := h.svc.AddRequest(c.Request().Context(), sd.UserID, req.Code, "", req.StartDate,req.EndDate,req.Unit,req.Quantity,req.Note)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, echo.Map{"data": request})

}


