package schedule

import (
	"time"
)

// DailyWorkCalculation holds computed work, overtime, and absence minutes.
type DailyWorkCalculation struct {
	WorkedMinutes   int
	OvertimeMinutes int
	AbsenceMinutes  int
}

// CalculateDailyWorkTime computes standard worked time, overtime, and absence based on planned vs raw worked minutes.
func CalculateDailyWorkTime(plannedMinutes, workedRawMinutes int) DailyWorkCalculation {
	if plannedMinutes <= 0 {
		return DailyWorkCalculation{
			WorkedMinutes:   0,
			OvertimeMinutes: workedRawMinutes,
			AbsenceMinutes:  0,
		}
	}

	if workedRawMinutes >= plannedMinutes {
		return DailyWorkCalculation{
			WorkedMinutes:   plannedMinutes,
			OvertimeMinutes: workedRawMinutes - plannedMinutes,
			AbsenceMinutes:  0,
		}
	}

	return DailyWorkCalculation{
		WorkedMinutes:   workedRawMinutes,
		OvertimeMinutes: 0,
		AbsenceMinutes:  plannedMinutes - workedRawMinutes,
	}
}

// DailyRestResult holds the result of validating Art. 132 KP 11-hour rest rule.
type DailyRestResult struct {
	RestDurationMinutes int
	RequiredMinutes     int
	DeficitMinutes      int
	Violated            bool
}

// ValidateDailyRest11h validates if the uninterrupted rest period between shifts satisfies Kodeks Pracy Art. 132 KP (11 hours = 660 minutes).
func ValidateDailyRest11h(previousEnd, currentStart time.Time) DailyRestResult {
	const requiredMinutes = 11 * 60 // 660 minut

	if currentStart.Before(previousEnd) {
		return DailyRestResult{
			RestDurationMinutes: 0,
			RequiredMinutes:     requiredMinutes,
			DeficitMinutes:      requiredMinutes,
			Violated:            true,
		}
	}

	restMinutes := int(currentStart.Sub(previousEnd).Minutes())
	if restMinutes < requiredMinutes {
		return DailyRestResult{
			RestDurationMinutes: restMinutes,
			RequiredMinutes:     requiredMinutes,
			DeficitMinutes:      requiredMinutes - restMinutes,
			Violated:            true,
		}
	}

	return DailyRestResult{
		RestDurationMinutes: restMinutes,
		RequiredMinutes:     requiredMinutes,
		DeficitMinutes:      0,
		Violated:            false,
	}
}

// CalculateNightMinutes calculates the minutes worked within the designated night window [nightStartHour..nightEndHour].
// Default Polish Labor Code window is typically 22:00 to 06:00 (spans midnight).
func CalculateNightMinutes(start, end time.Time, nightStartHour, nightEndHour int) int {
	if !end.After(start) {
		return 0
	}

	totalNightMinutes := 0
	// Iterate through each 24h day covering the interval
	dayCursor := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	if nightStartHour > nightEndHour {
		dayCursor = dayCursor.AddDate(0, 0, -1)
	}
	// Go up to the end time's day + 1 to account for shifts ending after midnight
	lastDay := time.Date(end.Year(), end.Month(), end.Day()+1, 0, 0, 0, 0, end.Location())

	for d := dayCursor; d.Before(lastDay); d = d.AddDate(0, 0, 1) {
		var winStart, winEnd time.Time
		if nightStartHour > nightEndHour {
			// Window spans midnight: e.g. 22:00 to 06:00
			// Window for day d: from d 22:00 to d+1 06:00
			winStart = time.Date(d.Year(), d.Month(), d.Day(), nightStartHour, 0, 0, 0, d.Location())
			winEnd = time.Date(d.Year(), d.Month(), d.Day()+1, nightEndHour, 0, 0, 0, d.Location())
		} else {
			// Window on same day: e.g. 00:00 to 06:00
			winStart = time.Date(d.Year(), d.Month(), d.Day(), nightStartHour, 0, 0, 0, d.Location())
			winEnd = time.Date(d.Year(), d.Month(), d.Day(), nightEndHour, 0, 0, 0, d.Location())
		}

		// Calculate overlap between [start, end] and [winStart, winEnd]
		overlapStart := start
		if winStart.After(overlapStart) {
			overlapStart = winStart
		}
		overlapEnd := end
		if winEnd.Before(overlapEnd) {
			overlapEnd = winEnd
		}

		if overlapEnd.After(overlapStart) {
			totalNightMinutes += int(overlapEnd.Sub(overlapStart).Minutes())
		}
	}

	return totalNightMinutes
}

// CalculateArt130NominalMinutes calculates statutory monthly nominal working minutes according to Art. 130 § 1 i § 2 Kodeksu Pracy:
// (40h * pełne_tygodnie) + (8h * dni_wystające_pn_pt) - (8h * święta_w_dni_inne_niż_niedziela)
func CalculateArt130NominalMinutes(year int, month time.Month, holidays []time.Time, ftePercent float64) int {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	totalDays := lastDay.Day()

	fullWeeks := totalDays / 7
	remainderDays := totalDays % 7

	// Liczba dni wystających przypadających od poniedziałku do piątku
	extraWorkingDays := 0
	for i := 0; i < remainderDays; i++ {
		checkDay := firstDay.AddDate(0, 0, fullWeeks*7+i)
		dow := checkDay.Weekday()
		if dow >= time.Monday && dow <= time.Friday {
			extraWorkingDays++
		}
	}

	nominalHours := (fullWeeks * 40) + (extraWorkingDays * 8)

	// Odejmujemy 8h za każde święto przypadające w okresie rozliczeniowym w innym dniu niż niedziela (art. 130 § 2 KP)
	holidayDeductions := 0
	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		if h.Year() == year && h.Month() == month {
			dateKey := h.Format("2006-01-02")
			if !holidayMap[dateKey] {
				holidayMap[dateKey] = true
				if h.Weekday() != time.Sunday {
					holidayDeductions++
				}
			}
		}
	}

	nominalHours -= (holidayDeductions * 8)
	if nominalHours < 0 {
		nominalHours = 0
	}

	if ftePercent <= 0 {
		ftePercent = 1.0
	}

	nominalMinutes := int(float64(nominalHours*60) * ftePercent)
	return nominalMinutes
}

// TardinessResult holds evaluation of grace period vs tardiness.
type TardinessResult struct {
	IsTardy        bool
	TardyMinutes   int
	WithinGrace    bool
	EffectiveStart time.Time
}

// EvaluateStartGrace evaluates punch in against planned start time and grace period in minutes.
func EvaluateStartGrace(plannedStart, actualPunchIn time.Time, graceMinutes int) TardinessResult {
	if !actualPunchIn.After(plannedStart) {
		// Arrived on time or earlier
		return TardinessResult{
			IsTardy:        false,
			TardyMinutes:   0,
			WithinGrace:    true,
			EffectiveStart: plannedStart,
		}
	}

	diffMinutes := int(actualPunchIn.Sub(plannedStart).Minutes())
	if diffMinutes <= graceMinutes {
		// Arrived slightly late but within allowed grace window
		return TardinessResult{
			IsTardy:        false,
			TardyMinutes:   0,
			WithinGrace:    true,
			EffectiveStart: plannedStart,
		}
	}

	// Tardy beyond grace period
	return TardinessResult{
		IsTardy:        true,
		TardyMinutes:   diffMinutes,
		WithinGrace:    false,
		EffectiveStart: actualPunchIn,
	}
}
