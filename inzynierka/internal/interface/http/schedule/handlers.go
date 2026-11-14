package schedulehttp

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/labstack/echo/v4"
	du "erpbadger/internal/domain/schedule"
	auth "erpbadger/internal/usecase/auth"
	uc "erpbadger/internal/usecase/schedule"
)

type Handlers struct {
	svc uc.Usecase
	authSvc auth.Usecase

}

func New(svc uc.Usecase, authSvc auth.Usecase) *Handlers { return &Handlers{svc: svc, authSvc: authSvc} }

func getUserID(c echo.Context) int {
	id, _ := strconv.Atoi(c.Param("id"))
	return id
}









// -------------------- REQUEST STRUCTS --------------------

type UpdateWeeklyRequest struct {
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	Days          []struct {
		DayOfWeek      int      `json:"dow"`
		PlannedMinutes int      `json:"planned_minutes"`
		EarliestStart  *string  `json:"earliest_start"`
		LatestStart    *string  `json:"latest_start"`
		LatestEnd      *string  `json:"latest_end"`
		Flex           bool     `json:"flex"`
		SpansMidnight  bool     `json:"spans_midnight"`
	} `json:"days"`
}


func (h *Handlers) DeleteTimePolicy(c echo.Context) error {
	idstring := c.Param("tpid")
	id,err := strconv.Atoi(idstring)

	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Nieprawidłowy ID polityki czasu pracy",
		})
	}
	derr := h.svc.DeleteTimePolicy(c.Request().Context(), id)
	if derr != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": derr.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Pomyślnie usunięto politykę czasu pracy",
	})



} 


func (h *Handlers) DeleteOverride(c echo.Context) error {
	idstring := c.Param("oid")
	id,err := strconv.Atoi(idstring)

	if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Nieprawidłowy ID dnia wyjątku",
		})
	}
	derr := h.svc.DeleteOverride(c.Request().Context(), id)
	if derr != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})

	}
	
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Pomyślnie usunięto dzień wyjątku",
	})



} 


// -------------------- RAPORT MIESIĘCZNY (PDF) --------------------
// -------------------- TIME POLICIES --------------------

// GET /api/time-policies
func (h *Handlers) ListTimePolicies(c echo.Context) error {
	policies, err := h.svc.ListTimePolicies(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	// Zwracamy strukturę domenową du.TimePolicy:
	// - EffectiveFrom / EffectiveTo -> time.Time / *time.Time (JSON = RFC3339)
	// - EarliestStart / LatestStart / LatestEnd -> *string "HH:MM"
	return c.JSON(http.StatusOK, echo.Map{
		"data": policies,
	})
}
// POST /api/time-policies
func (h *Handlers) SaveTimePolicy(c echo.Context) error {
	var req du.TimePolicyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Błędne dane (JSON)",
		})
	}

	if strings.TrimSpace(req.EffectiveFrom) == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Pole effective_from jest wymagane (YYYY-MM-DD)",
		})
	}

	// parsujemy daty
	effFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Nieprawidłowy format effective_from, oczekiwano YYYY-MM-DD",
		})
	}

	var effTo *time.Time
	if req.EffectiveTo != nil && strings.TrimSpace(*req.EffectiveTo) != "" {
		t, err := time.Parse("2006-01-02", *req.EffectiveTo)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "Nieprawidłowy format effective_to, oczekiwano YYYY-MM-DD",
			})
		}
		effTo = &t
	}

	p := du.TimePolicy{
		ID:            req.ID,
		EffectiveFrom: effFrom,
		EffectiveTo:   effTo,

		EarliestStart: req.EarliestStart, // "HH:MM" lub nil
		LatestStart:   req.LatestStart,
		LatestEnd:     req.LatestEnd,

		Flex: req.Flex,
	}

	if err := h.svc.SaveTimePolicy(c.Request().Context(), p); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	// zwracamy zapisany obiekt – front może odświeżyć listę albo użyć danych
	return c.JSON(http.StatusOK, echo.Map{
		"data": p,
	})
}
// GET /api/time-policies/active?date=YYYY-MM-DD
func (h *Handlers) GetActiveTimePolicy(c echo.Context) error {
	dateStr := c.QueryParam("date")
	if strings.TrimSpace(dateStr) == "" {
		// jak nie podano, bierzemy "dzisiaj"
		dateStr = time.Now().Format("2006-01-02")
	}

	day, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "Nieprawidłowy format date, oczekiwano YYYY-MM-DD",
		})
	}

	p, err := h.svc.GetActiveTimePolicy(c.Request().Context(), day)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	// jeśli brak polityki – data: null
	return c.JSON(http.StatusOK, echo.Map{
		"data": p,
	})
}

func (h *Handlers) MonthlyReport(c echo.Context) error {
	userIDStr := c.Param("user_id")



	month := c.Param("month") // "2025-12"

	if userIDStr == "" || month == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "nieprawidłowe parametry raportu (user_id, month)",
		})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "user_id musi być liczbą > 0",
		})
	}





	scope,_, err := h.authSvc.MaxScopeFor(c.Request().Context(), "schedule:read")
	if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Nieudana próba pobrania uprawnień",
				})
	}

    	switch scope {
		case "own":

			is,_ := h.authSvc.CheckIsYourself(c.Request().Context(),userID)
			if is {
			} else {
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})
			}

		case "team":
			is,_ := h.authSvc.CheckIsUserInMyDepartment(c.Request().Context(),userID)
			if is {
			} else {
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})
			}


		case "org":

		default:
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})

			}










	pdfOut, err := h.svc.GenerateMonthlyReportPDF(c.Request().Context(), userID, month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "błąd budowania PDF",
			"details": err.Error(),
		})
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%q", filepath.Base(pdfOut)))

	return c.File(pdfOut)
}
// -------------------- WEEKLY --------------------

// POST /api/user/:id/schedule/weekly/:week (zakładanie nowego grafiku)
func (h *Handlers) NewWeekly(c echo.Context) error {
	week := c.Param("week") // np. "2024-01-01|2024-02-01"
	if week == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "nieprawidłowy parametr tygodnia",
		})
	}

	parts := strings.SplitN(week, "|", 2)

	var from, to *string

	if len(parts) > 0 {
		v := strings.TrimSpace(parts[0])
		if v != "" && v != "null" {
			from = new(string)
			*from = v
		}
	}

	if len(parts) > 1 {
		v := strings.TrimSpace(parts[1])
		if v != "" && v != "null" {
			to = new(string)
			*to = v
		}
	}

	data, err := h.svc.NewUserSchedule(c.Request().Context(), getUserID(c), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}



func (h *Handlers) DeleteWeekly(c echo.Context) error {
	week := c.Param("week") // np. "2024-01-01|2024-02-01"
	if week == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "nieprawidłowy parametr tygodnia",
		})
	}

	parts := strings.SplitN(week, "|", 2)

	var from string
	var to *string

	if len(parts) > 0 {
		v := strings.TrimSpace(parts[0])
		from = v
	}

	if len(parts) > 1 {
		v := strings.TrimSpace(parts[1])
		if v != "" && v != "null" {
			to = new(string)
			*to = v
		}
	}

	err := h.svc.DeleteWeekly(c.Request().Context(), getUserID(c), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Usunięto plan tygodniowy"})
}






// GET /api/user/:id/schedule/weekly/:week
func (h *Handlers) GetWeekly(c echo.Context) error {
	week := c.Param("week")
	if week == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "nieprawidłowy parametr tygodnia",
		})
	}

	parts := strings.SplitN(week, "|", 2)

	var from, to *string

	if len(parts) > 0 {
		v := strings.TrimSpace(parts[0])
		if v != "" && v != "null" {
			from = new(string)
			*from = v
		}
	}

	if len(parts) > 1 {
		v := strings.TrimSpace(parts[1])
		if v != "" && v != "null" {
			to = new(string)
			*to = v
		}
	}

	data, err := h.svc.GetUserSchedule(c.Request().Context(), getUserID(c), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// GET /api/user/:id/schedule/weeks
func (h *Handlers) GetWeeks(c echo.Context) error {
	data, err := h.svc.GetUserScheduleWeeks(c.Request().Context(), getUserID(c))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// PUT /api/user/:id/schedule/weekly/:week
func (h *Handlers) UpdateWeekly(c echo.Context) error {
	week := c.Param("week")
	if week == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "nieprawidłowy parametr tygodnia",
		})
	}

	parts := strings.SplitN(week, "|", 2)



	var to *string


	if len(parts) > 1 {
		v := strings.TrimSpace(parts[1])
		if v != "" && v != "null" {
			to = new(string)
			*to = v
		}
	}

	var req UpdateWeeklyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Błędne dane (JSON)"})
	}
effectiveFrom := ""
if req.EffectiveFrom != nil {
	effectiveFrom = *req.EffectiveFrom
}
	var s []du.UserSchedule
	for _, d := range req.Days {
		s = append(s, du.UserSchedule{
			DayOfWeek:      d.DayOfWeek,
			PlannedMinutes: d.PlannedMinutes,
			EffectiveFrom:  req.EffectiveFrom,
			EffectiveTo:    req.EffectiveTo,
			EarliestStart:  d.EarliestStart,
			LatestStart:    d.LatestStart,
			LatestEnd:      d.LatestEnd,
			Flex:           d.Flex,
			SpansMidnight:  d.SpansMidnight,
		})
	}

	if err := h.svc.UpdateUserSchedule(c.Request().Context(), getUserID(c), s, strings.TrimSpace(parts[0]), to, effectiveFrom, req.EffectiveTo); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": "Grafik zaktualizowany"})
}

// -------------------- OVERRIDES --------------------

// GET /api/user/:id/schedule/overrides?month=YYYY-MM
func (h *Handlers) ListOverrides(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("id"))
	q := c.QueryParam("month") // np. "2025-11"

	data, err := h.svc.ListOverrides(c.Request().Context(), userID, q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// POST /api/user/:id/schedule/overrides
func (h *Handlers) AddOverride(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Day            string  `json:"day"`
		PlannedMinutes int     `json:"planned_minutes"`
		EarliestStart  *string `json:"earliest_start"`
		LatestStart    *string `json:"latest_start"`
		LatestEnd      *string `json:"latest_end"`
		Flex           bool    `json:"flex"`
		SpansMidnight  bool    `json:"spans_midnight"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Błędne dane (JSON)"})
	}

	if req.Day == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Brak daty dnia (day)"})
	}

	err := h.svc.AddOverride(
		c.Request().Context(),
		userID,
		req.Day,
		req.PlannedMinutes,
		req.EarliestStart,
		req.LatestStart,
		req.LatestEnd,
		req.Flex,
		req.SpansMidnight,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": "Zaktualizowano wyjątek grafiku"})
}

// -------------------- MIESIĘCZNE (JSON) --------------------

// GET /api/user/:userId/schedule/monthly/:month
func (h *Handlers) GetMonthly(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))
	month := c.Param("month")

	if month == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing month parameter"})
	}

	data, err := h.svc.GetMonthly(c.Request().Context(), userID, month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// GET /api/schedule/monthly/users?month=YYYY-MM[&departments=1,3]
func (h *Handlers) GetMonthlyForUsers(c echo.Context) error {
	month := c.QueryParam("month")
	if month == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "missing month parameter (expected YYYY-MM)",
		})
	}

	deptParam := c.QueryParam("departments")
	var deptIDs []int
	if deptParam != "" {
		parts := strings.Split(deptParam, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := strconv.Atoi(p)
			if err != nil {
				return c.JSON(http.StatusBadRequest, echo.Map{
					"error": "invalid departments parameter – expected numeric ids",
				})
			}
			deptIDs = append(deptIDs, id)
		}
	}

	data, err := h.svc.GetMonthlyForAllEmployees(c.Request().Context(), month, deptIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// GET /api/schedule/monthly/all?month=YYYY-MM[&departments=1,3]
func (h *Handlers) GetMonthlyForAllUsers(c echo.Context) error {
	month := c.QueryParam("month")
	if month == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "missing month parameter (expected YYYY-MM)",
		})
	}

	re := regexp.MustCompile(`^\d{4}-\d{2}$`)
	if !re.MatchString(month) {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid month format, expected YYYY-MM",
		})
	}

	deptParam := c.QueryParam("departments")
	var deptIDs []int
	if deptParam != "" {
		parts := strings.Split(deptParam, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := strconv.Atoi(p)
			if err != nil {
				return c.JSON(http.StatusBadRequest, echo.Map{
					"error": "invalid departments parameter – expected numeric ids",
				})
			}
			deptIDs = append(deptIDs, id)
		}
	}

	data, err := h.svc.GetMonthlyForAllEmployees(c.Request().Context(), month, deptIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": data})
}











// GET /api/user/:userId/schedule/monthly/:month
func (h *Handlers) GetMonthlyNew(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))
	month := c.Param("month")

	scope,_, err := h.authSvc.MaxScopeFor(c.Request().Context(), "schedule:read")
	if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Nieudana próba pobrania uprwnień",
				})
	}

    	switch scope {
		case "own":

			is,_ := h.authSvc.CheckIsYourself(c.Request().Context(),userID)
			if is {
			} else {
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})
			}

		case "team":
			is,_ := h.authSvc.CheckIsUserInMyDepartment(c.Request().Context(),userID)
			if is {
			} else {
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})
			}


		case "org":

		default:
					return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "Brak Uprawnień",
				})

			}




	if month == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "brak parametru miesiąc"})
	}

	data, err := h.svc.GetMonthlyNew(c.Request().Context(), userID, month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

