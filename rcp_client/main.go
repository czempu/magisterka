package main

import (
	_ "embed"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed index.html
var indexHTML []byte

type ShiftDefinition struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	StartHour     int    `json:"start_hour"`
	DurationHours int    `json:"duration_hours"`
}

type EmployeePreference struct {
	PrefDate         *string `json:"pref_date,omitempty"`
	PreferredShiftID *int    `json:"preferred_shift_id,omitempty"`
	WantFreeDay      bool    `json:"want_free_day"`
	Weight           int     `json:"weight"`
}

type EmployeeProfile struct {
	ID                 int                  `json:"id"`
	Name               string               `json:"name"`
	ContractFTE        float64              `json:"contract_fte"`
	Skills             []string             `json:"skills"`
	DisallowedShiftIDs []int                `json:"disallowed_shift_ids"`
	VacationDays       []string             `json:"vacation_days"`
	Preferences        []EmployeePreference `json:"preferences"`
}

type ScheduleRequest struct {
	Employees                      []EmployeeProfile `json:"employees"`
	StartDate                      string            `json:"start_date"`
	EndDate                        string            `json:"end_date"`
	Shifts                         []ShiftDefinition `json:"shifts,omitempty"`
	RequireSundayRestEvery4Weeks   bool              `json:"require_sunday_rest_every_4_weeks"`
	EnforceForwardRotation         bool              `json:"enforce_forward_rotation"`
	FairnessBalanceNights          bool              `json:"fairness_balance_nights"`
	MaxConsecutiveDays             *int              `json:"max_consecutive_days,omitempty"`
	AutoCalculateWorkNorm          bool              `json:"auto_calculate_work_norm"`
	MaxWorkHours                   *int              `json:"max_work_hours,omitempty"`
	CustomHolidays                 []string          `json:"custom_holidays"`
	ClosedDays                     []string          `json:"closed_days"`
	Demand                         map[string]int    `json:"demand"`
	SkillDemand                    map[string]int    `json:"skill_demand"`
	SolverChoice                   string            `json:"solver_choice"`
	TimeLimitSeconds               float64           `json:"time_limit_seconds"`
}

type TaskSubmissionResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message"`
}

type ScheduleResult struct {
	SolverName              string              `json:"solver_name"`
	Status                  string              `json:"status"`
	SolveTimeSeconds        float64             `json:"solve_time_seconds"`
	ObjectiveValue          *float64            `json:"objective_value"`
	Schedule                map[string]int      `json:"schedule"`
	AssignedHours           map[string]int      `json:"assigned_hours"`
	EmployeeNames           map[string]string   `json:"employee_names"`
	EmployeeSkills          map[string][]string `json:"employee_skills"`
	ShiftDefinitions        []ShiftDefinition   `json:"shift_definitions"`
	CalendarDates           []string            `json:"calendar_dates"`
	CalculatedWorkNormHours *int                `json:"calculated_work_norm_hours"`
	Holidays                []string            `json:"holidays"`
	ClosedDays              []string            `json:"closed_days"`
}

type TaskStatusResponse struct {
	TaskID               string          `json:"task_id"`
	Status               string          `json:"status"`
	SolverChoice         string          `json:"solver_choice"`
	CreatedAt            string          `json:"created_at"`
	ExecutionTimeSeconds *float64        `json:"execution_time_seconds"`
	Result               *ScheduleResult `json:"result"`
	ErrorMessage         *string         `json:"error_message"`
}

var schedulerURL = "http://localhost:8000"

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda niedozwolona", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Błąd odczytu żądania", http.StatusBadRequest)
		return
	}

	// 1. Zlecenie zadania do Pythona (POST /api/v1/schedules/submit)
	submitURL := fmt.Sprintf("%s/api/v1/schedules/submit", schedulerURL)
	resp, err := http.Post(submitURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Błąd połączenia z mikroserwisem solvera: %v", err)
		http.Error(w, fmt.Sprintf("Błąd połączenia z mikroserwisem: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Solver zwrócił błąd (%d): %s", resp.StatusCode, string(respBody)), http.StatusInternalServerError)
		return
	}

	var submission TaskSubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&submission); err != nil {
		http.Error(w, "Błąd parsowania odpowiedzi submission", http.StatusInternalServerError)
		return
	}

	// 2. Polling statusu (GET /api/v1/schedules/tasks/{id})
	timeout := time.After(35 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			http.Error(w, "Przekroczono limit czasu oczekiwania na solver (timeout)", http.StatusGatewayTimeout)
			return
		case <-ticker.C:
			statusURL := fmt.Sprintf("%s/api/v1/schedules/tasks/%s", schedulerURL, submission.TaskID)
			sResp, err := http.Get(statusURL)
			if err != nil {
				continue
			}

			var taskStatus TaskStatusResponse
			err = json.NewDecoder(sResp.Body).Decode(&taskStatus)
			sResp.Body.Close()
			if err != nil {
				continue
			}

			if taskStatus.Status == "COMPLETED" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(taskStatus.Result)
				return
			} else if taskStatus.Status == "FAILED" {
				errMsg := "Brak szczegółów błędu"
				if taskStatus.ErrorMessage != nil {
					errMsg = *taskStatus.ErrorMessage
				}
				http.Error(w, fmt.Sprintf("Optymalizacja nie powiodła się: %s", errMsg), http.StatusInternalServerError)
				return
			}
		}
	}
}

func handleExportExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda niedozwolona", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Błąd odczytu żądania", http.StatusBadRequest)
		return
	}

	excelURL := fmt.Sprintf("%s/api/v1/schedules/export-excel", schedulerURL)
	resp, err := http.Post(excelURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Błąd mikroserwisu: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=harmonogram_pracy.xlsx")
	io.Copy(w, resp.Body)
}

func handleBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda niedozwolona", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Błąd odczytu żądania", http.StatusBadRequest)
		return
	}

	benchURL := fmt.Sprintf("%s/api/v1/schedules/benchmark", schedulerURL)
	resp, err := http.Post(benchURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Błąd połączenia z solverem: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	if url := os.Getenv("SCHEDULER_URL"); url != "" {
		schedulerURL = url
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/export-excel", handleExportExcel)
	mux.HandleFunc("/api/benchmark", handleBenchmark)

	log.Printf("==================================================================")
	log.Printf("   RCP CLIENT & DASHBOARD (Go Web Core)")
	log.Printf("   Adres Web UI: http://localhost:%s", port)
	log.Printf("   Silnik Optymalizacji (Python): %s", schedulerURL)
	log.Printf("==================================================================")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Błąd serwera HTTP: %v", err)
	}
}
