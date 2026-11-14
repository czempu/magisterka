package schedule_test

import (
	"testing"
	"time"

	"erpbadger/internal/domain/schedule"
)

func TestCalculateDailyWorkTime(t *testing.T) {
	tests := []struct {
		name            string
		planned         int
		workedRaw       int
		wantWorked      int
		wantOvertime    int
		wantAbsence     int
	}{
		{
			name:         "Exact standard 8h shift",
			planned:      480,
			workedRaw:    480,
			wantWorked:   480,
			wantOvertime: 0,
			wantAbsence:  0,
		},
		{
			name:         "Overtime on working day (10h worked)",
			planned:      480,
			workedRaw:    600,
			wantWorked:   480,
			wantOvertime: 120,
			wantAbsence:  0,
		},
		{
			name:         "Underworked on working day (6h worked)",
			planned:      480,
			workedRaw:    360,
			wantWorked:   360,
			wantOvertime: 0,
			wantAbsence:  120,
		},
		{
			name:         "Work on day off / weekend (0h planned, 5h worked)",
			planned:      0,
			workedRaw:    300,
			wantWorked:   0,
			wantOvertime: 300,
			wantAbsence:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schedule.CalculateDailyWorkTime(tt.planned, tt.workedRaw)
			if got.WorkedMinutes != tt.wantWorked {
				t.Errorf("WorkedMinutes = %d, want %d", got.WorkedMinutes, tt.wantWorked)
			}
			if got.OvertimeMinutes != tt.wantOvertime {
				t.Errorf("OvertimeMinutes = %d, want %d", got.OvertimeMinutes, tt.wantOvertime)
			}
			if got.AbsenceMinutes != tt.wantAbsence {
				t.Errorf("AbsenceMinutes = %d, want %d", got.AbsenceMinutes, tt.wantAbsence)
			}
		})
	}
}

func TestValidateDailyRest11h(t *testing.T) {
	loc := time.UTC
	prevEnd := time.Date(2026, 10, 1, 16, 0, 0, 0, loc)

	// Case 1: Sufficient rest (16 hours rest)
	nextStartGood := time.Date(2026, 10, 2, 8, 0, 0, 0, loc)
	resGood := schedule.ValidateDailyRest11h(prevEnd, nextStartGood)
	if resGood.Violated {
		t.Errorf("expected no violation for 16h rest, got violated with deficit %d", resGood.DeficitMinutes)
	}
	if resGood.RestDurationMinutes != 16*60 {
		t.Errorf("expected 960 min rest, got %d", resGood.RestDurationMinutes)
	}

	// Case 2: Violated rest (only 8 hours rest between 22:00 and 06:00)
	lateShiftEnd := time.Date(2026, 10, 1, 22, 0, 0, 0, loc)
	earlyNextStart := time.Date(2026, 10, 2, 6, 0, 0, 0, loc)
	resViolated := schedule.ValidateDailyRest11h(lateShiftEnd, earlyNextStart)
	if !resViolated.Violated {
		t.Errorf("expected violation for 8h rest, got valid")
	}
	if resViolated.DeficitMinutes != 180 { // 11h - 8h = 3h = 180 min
		t.Errorf("expected deficit 180 min, got %d", resViolated.DeficitMinutes)
	}
}

func TestCalculateNightMinutes(t *testing.T) {
	loc := time.UTC

	// Daytime shift: 08:00 - 16:00 -> 0 night min
	s1 := time.Date(2026, 10, 1, 8, 0, 0, 0, loc)
	e1 := time.Date(2026, 10, 1, 16, 0, 0, 0, loc)
	if got := schedule.CalculateNightMinutes(s1, e1, 22, 6); got != 0 {
		t.Errorf("daytime shift: expected 0 night min, got %d", got)
	}

	// Night shift: 22:00 - 06:00 (full night shift) -> 480 min
	s2 := time.Date(2026, 10, 1, 22, 0, 0, 0, loc)
	e2 := time.Date(2026, 10, 2, 6, 0, 0, 0, loc)
	if got := schedule.CalculateNightMinutes(s2, e2, 22, 6); got != 480 {
		t.Errorf("full night shift: expected 480 night min, got %d", got)
	}

	// Partial evening shift: 18:00 - 23:30 -> 90 night min (22:00 - 23:30)
	s3 := time.Date(2026, 10, 1, 18, 0, 0, 0, loc)
	e3 := time.Date(2026, 10, 1, 23, 30, 0, 0, loc)
	if got := schedule.CalculateNightMinutes(s3, e3, 22, 6); got != 90 {
		t.Errorf("partial evening shift: expected 90 night min, got %d", got)
	}

	// Early morning shift: 04:00 - 12:00 -> 120 night min (04:00 - 06:00)
	s4 := time.Date(2026, 10, 2, 4, 0, 0, 0, loc)
	e4 := time.Date(2026, 10, 2, 12, 0, 0, 0, loc)
	if got := schedule.CalculateNightMinutes(s4, e4, 22, 6); got != 120 {
		t.Errorf("early morning shift: expected 120 night min, got %d", got)
	}
}

func TestCalculateArt130NominalMinutes(t *testing.T) {
	// October 2026: 31 days. Oct 1 is Thursday.
	// Weeks: 4 full weeks (28 days) = 160h.
	// Extra days: 3 (Thu Oct 29, Fri Oct 30, Sat Oct 31). Extra working days (Mon-Fri) = 2 -> 16h.
	// Total nominal hours = 176h = 10560 minutes.
	minutesOct2026 := schedule.CalculateArt130NominalMinutes(2026, time.October, nil, 1.0)
	expectedOct := 176 * 60
	if minutesOct2026 != expectedOct {
		t.Errorf("Oct 2026: expected %d min (%d h), got %d min", expectedOct, 176, minutesOct2026)
	}

	// November 2026: Nov 1 (All Saints) is Sunday -> no deduction.
	// Nov 11 (Independence Day) is Wednesday -> falls on weekday, deducts 8h!
	nov11 := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	nov1 := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	holidaysNov := []time.Time{nov1, nov11}

	// Nov has 30 days. Nov 1 is Sunday.
	// 4 full weeks (28 days) = 160h.
	// Extra days: 2 (Sun Nov 29, Mon Nov 30). Mon is weekday -> 8h.
	// Gross = 168h. Deduction for Nov 11 (Wed) = 8h. Total = 160h = 9600 min.
	minutesNov2026 := schedule.CalculateArt130NominalMinutes(2026, time.November, holidaysNov, 1.0)
	expectedNov := 160 * 60
	if minutesNov2026 != expectedNov {
		t.Errorf("Nov 2026 with holiday deduction: expected %d min, got %d min", expectedNov, minutesNov2026)
	}

	// Test half-time employee (FTE 0.5)
	minutesHalfTime := schedule.CalculateArt130NominalMinutes(2026, time.November, holidaysNov, 0.5)
	if minutesHalfTime != expectedNov/2 {
		t.Errorf("Nov 2026 half-time: expected %d min, got %d min", expectedNov/2, minutesHalfTime)
	}
}

func TestEvaluateStartGrace(t *testing.T) {
	loc := time.UTC
	planned := time.Date(2026, 10, 1, 8, 0, 0, 0, loc)

	// Punch in on time
	onTime := time.Date(2026, 10, 1, 8, 0, 0, 0, loc)
	res1 := schedule.EvaluateStartGrace(planned, onTime, 5)
	if res1.IsTardy || !res1.WithinGrace {
		t.Errorf("expected not tardy for on-time punch")
	}

	// Punch in 4 minutes late with 5 min grace
	withinGrace := time.Date(2026, 10, 1, 8, 4, 0, 0, loc)
	res2 := schedule.EvaluateStartGrace(planned, withinGrace, 5)
	if res2.IsTardy || !res2.WithinGrace {
		t.Errorf("expected within grace for 4 min late with 5 min grace")
	}

	// Punch in 12 minutes late with 5 min grace
	tardy := time.Date(2026, 10, 1, 8, 12, 0, 0, loc)
	res3 := schedule.EvaluateStartGrace(planned, tardy, 5)
	if !res3.IsTardy || res3.WithinGrace || res3.TardyMinutes != 12 {
		t.Errorf("expected tardy 12 min, got %+v", res3)
	}
}
