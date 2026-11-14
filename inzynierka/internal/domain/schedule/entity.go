package schedule

import (
	"time"
	"encoding/json"

)

// -------------------- GRAFIK TYGODNIOWY --------------------

type UserSchedule struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	DayOfWeek      int     `json:"dow"`
	PlannedMinutes int     `json:"planned_minutes"`

	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`

	// Okno startu i końca dnia pracy na poziomie użytkownika (grafik),
	// analogicznie do time_policies:
	// - earliest_start / latest_start: przedział dopuszczalnego startu dnia,
	// - latest_end: maksymalny koniec dnia (w ramach tej doby / polityki),
	// - flex: true -> dzień elastyczny (wymiar + okno, bez sztywnej godziny),
	// - spans_midnight: true -> zmiana przechodzi przez północ.
	EarliestStart *string `json:"earliest_start,omitempty"` // "HH:MM"
	LatestStart   *string `json:"latest_start,omitempty"`   // "HH:MM"
	LatestEnd     *string `json:"latest_end,omitempty"`     // "HH:MM"

	Flex          bool `json:"flex"`
	SpansMidnight bool `json:"spans_midnight"`
}

// Agregacja tygodniówek po zakresach effective_from / effective_to
type UserScheduleWeeks struct {
	ID            int     `json:"id"` // faktycznie user_id, ale nie ruszamy nazwy
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
}

// -------------------- OVERRIDES (WYJĄTKI DZIENNE) --------------------

type UserScheduleOverride struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Day            time.Time `json:"day"`
	PlannedMinutes int       `json:"planned_minutes"`

	EarliestStart *string `json:"earliest_start,omitempty"`
	LatestStart   *string `json:"latest_start,omitempty"`
	LatestEnd     *string `json:"latest_end,omitempty"`

	Flex          bool `json:"flex"`
	SpansMidnight bool `json:"spans_midnight"`
}

// -------------------- RAPORT MIESIĘCZNY --------------------

type DailyEntry struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Type  string `json:"type"`
}

type DailyRecord struct {
	Date           string       `json:"date"`
	Weekday        int          `json:"weekday"`
	PlannedMinutes int          `json:"planned_minutes"`
	PlanSource     string       `json:"plan_source"`
	LeaveMinutes   int          `json:"leave_minutes"`
	LeaveType      string       `json:"leave_type"`
	WorkedMinutes  int          `json:"worked_minutes"`
	BreakMinutes   int          `json:"break_minutes"`
	Overtime       int          `json:"overtime_minutes"`
	Entries        []DailyEntry `json:"entries"`
}

type MonthlySchedule struct {
	UserID int           `json:"user_id"`
	Month  string        `json:"month"`
	Days   []DailyRecord `json:"days"`
}

type CompactDailyRecord struct {
	Date           string `json:"date"`
	Weekday        int    `json:"weekday"`
	PlannedMinutes int    `json:"planned_minutes"`
	PlanSource     string `json:"plan_source"`
	LeaveMinutes   int    `json:"leave_minutes"`
	LeaveType      string `json:"leave_type"`
	WorkedMinutes  int    `json:"worked_minutes"`
	BreakMinutes   int    `json:"break_minutes"`
	Overtime       int    `json:"overtime_minutes"`
}

type CompactMonthlySchedule struct {
	UserID int                  `json:"user_id"`
	Month  string               `json:"month"`
	Days   []CompactDailyRecord `json:"days"`
}

type MultiUserMonthlySchedule struct {
	Month string                   `json:"month"`
	Users []CompactMonthlySchedule `json:"users"`
}

// -------------------- TIME POLICY (POZIOM FIRMY) --------------------

// Firmowa polityka czasu pracy w danym okresie.
type TimePolicy struct {
	ID int64 `json:"id"`

	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`

	EarliestStart *string `json:"earliest_start,omitempty"` // "HH:MM"
	LatestStart   *string `json:"latest_start,omitempty"`   // "HH:MM"
	LatestEnd     *string `json:"latest_end,omitempty"`     // "HH:MM"

	Flex bool `json:"flex"`
}
type TimePolicyRequest struct {
	ID            int64   `json:"id"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"` // null albo "" = brak końca

	EarliestStart *string `json:"earliest_start"` // "HH:MM" lub null
	LatestStart   *string `json:"latest_start"`   // "HH:MM" lub null
	LatestEnd     *string `json:"latest_end"`     // "HH:MM" lub null

	Flex bool `json:"flex"`
}
type MonthlyReport struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Surname string `json:"surname"`

	FromDate string `json:"from_date"` // YYYY-MM-DD
	ToDate   string `json:"to_date"`   // YYYY-MM-DD

	Totals MonthlyTotals `json:"totals"`
	Days   []DayReportRow `json:"days"`
}
type MonthlyTotals struct {
	MonthPlanMinutes           int `json:"month_plan_minutes"`
	MonthExecMinutes           int `json:"month_exec_minutes"`
	MonthNetDeltaMinutes       int `json:"month_net_delta_minutes"`
	MonthTargetMinutes         int `json:"month_target_minutes"`
	MonthExecMinusPlanMinutes  int `json:"month_exec_minus_plan_minutes"`
	MonthExecMinusTargetMinutes int `json:"month_exec_minus_target_minutes"`
	NominalMissingAny          int `json:"nominal_missing_any"`
}

type DayReportRow struct {
	Day    string `json:"day"` // w JSON będzie RFC3339; jeśli chcesz YYYY-MM-DD, sformatuj w handlerze/DTO
	UserID int       `json:"user_id"`

	Name    string `json:"name"`
	Surname string `json:"surname"`

	Dow     int    `json:"dow"`
	DowName string `json:"dow_name"`

	PlanSource    string `json:"plan_source"`
	PlanType      string `json:"plan_type"`
	ScheduleHours string `json:"schedule_hours"`

	PlanMinutesRaw int `json:"plan_minutes_raw"`
	PlanMinutes    int `json:"plan_minutes"`
	RequiredMinutes int `json:"required_minutes"`

	LeaveTypeCodes     *string    `json:"leave_type_codes,omitempty"`
	LeaveNames         *string    `json:"leave_names,omitempty"`
	LeaveMinutesRaw    int        `json:"leave_minutes_raw"`
	LeaveCreditMinutes int        `json:"leave_credit_minutes"`
	LeaveFirstStartAt  *string `json:"leave_first_start_at,omitempty"`
	LeaveLastEndAt     *string `json:"leave_last_end_at,omitempty"`

	SickID              *int    `json:"sick_id,omitempty"`
	LegalBasis          *string `json:"legal_basis,omitempty"`
	SickWorkAccident    int     `json:"sick_work_accident"`
	SickPercentCoverage int     `json:"sick_percent_coverage"`

	HolidayName *string `json:"holiday_name,omitempty"`

	WorkMinutes  int `json:"work_minutes"`
	BreakMinutes int `json:"break_minutes"`
	NightMinutes int `json:"night_minutes"`

	RealStart *string `json:"real_start,omitempty"` // TIME_FORMAT z SQL
	RealEnd   *string `json:"real_end,omitempty"`   // TIME_FORMAT z SQL

	BankUsageMinutes int `json:"bank_usage_minutes"`
	BankUsedMinutes  int `json:"bank_used_minutes"`
	PayoutMinutes    int `json:"payout_minutes"`
	ToLeaveMinutes   int `json:"to_leave_minutes"`

	CreditedToNormMinutes int `json:"credited_to_norm_minutes"`
	GrossDeltaMinutes     int `json:"gross_delta_minutes"`
	NetDeltaMinutes       int `json:"net_delta_minutes"`

	LateMinutes       int `json:"late_minutes"`
	EarlyLeaveMinutes int `json:"early_leave_minutes"`

	RestMinutes *int   `json:"rest_minutes,omitempty"`
	RestAlert   string `json:"rest_alert"`
	Status      string `json:"status"`

	FtePercent       float64 `json:"fte_percent"`
	NominalMinutes100 float64 `json:"nominal_minutes_100"`
	NominalMissing   int     `json:"nominal_missing"`

	TargetShareMinutes int `json:"target_share_minutes"`

	CumTargetMinutes          int `json:"cum_target_minutes"`
	CumPlanMinutes            int `json:"cum_plan_minutes"`
	CumExecMinutes            int `json:"cum_exec_minutes"`
	CumNetDeltaMinutes        int `json:"cum_net_delta_minutes"`
	CumExecMinusPlanMinutes   int `json:"cum_exec_minus_plan_minutes"`
	CumExecMinusTargetMinutes int `json:"cum_exec_minus_target_minutes"`

	// surowy JSON z MySQL (string -> RawMessage). Trzymamy jako RawMessage, żeby nie parsować w repo.
	InOut json.RawMessage `json:"in_out,omitempty"`
}
