package schedule

import (
	"fmt"
	"context"
	"time"
	"errors"
	du "erpbadger/internal/domain/schedule"
)

type Usecase interface {
	// WEEKLY
	GetUserSchedule(ctx context.Context, userID int, from, to *string) ([]du.UserSchedule, error)
	GetUserScheduleWeeks(ctx context.Context, userID int) ([]du.UserScheduleWeeks, error)
	NewUserSchedule(ctx context.Context, userID int, from, to *string) (string, error)
	UpdateUserSchedule(ctx context.Context, userID int, s []du.UserSchedule, from string, to *string,oldFrom string, oldTo *string) error

	// OVERRIDES
	ListOverrides(ctx context.Context, userID int, month string) ([]du.UserScheduleOverride, error)
	AddOverride(ctx context.Context, userID int, day string, planned int, earliestStart, latestStart, latestEnd *string, flex, spans bool) error

	// RAPORTY MIESIĘCZNE
	GetMonthly(ctx context.Context, userID int, month string) (du.MonthlySchedule, error)
	GetMonthlyForUsers(ctx context.Context, userIDs []int, month string) (du.MultiUserMonthlySchedule, error)
	GetMonthlyForAllEmployees(ctx context.Context, month string, departmentIDs []int) (du.MultiUserMonthlySchedule, error)
	GetMonthlyNew(ctx context.Context, userID int, month string) (du.MonthlyReport, error)
	GenerateMonthlyReportPDF(ctx context.Context, userID int, month string) (string, error)

	// TIME POLICIES
	GetActiveTimePolicy(ctx context.Context, day time.Time) (*du.TimePolicy, error)
	ListTimePolicies(ctx context.Context) ([]du.TimePolicy, error)
	SaveTimePolicy(ctx context.Context, p du.TimePolicy) error

	DeleteTimePolicy(ctx context.Context, id int) error
	DeleteOverride(ctx context.Context, id int) error
	DeleteWeekly(ctx context.Context, userID int, from string, to *string) error
}

type usecase struct {
	repo      du.Repository
	reportGen du.ReportGenerator
}

func NewUsecase(r du.Repository, reportGen du.ReportGenerator) Usecase {
	return &usecase{repo: r, reportGen: reportGen}
}


func (u *usecase) DeleteOverride(ctx context.Context, id int) error {
	return u.repo.DeleteOverride(ctx, id)
}



func (u *usecase) DeleteTimePolicy(ctx context.Context, id int) error {
	return u.repo.DeleteTimePolicies(ctx, id)
}




// -------------------- WEEKLY --------------------

func (u *usecase) GetUserScheduleWeeks(ctx context.Context, userID int) ([]du.UserScheduleWeeks, error) {
	return u.repo.GetWeeks(ctx, userID)
}

func (u *usecase) GetUserSchedule(ctx context.Context, userID int, from, to *string) ([]du.UserSchedule, error) {
	return u.repo.GetWeeklySchedule(ctx, userID, from, to)
}

func (u *usecase) NewUserSchedule(ctx context.Context, userID int, from, to *string) (string, error) {
	_, err := u.repo.CheckWeekSchedule(ctx, userID, from, to)
	if err != nil {
		return "błąd", err
	}
	return "dodano", nil
}

func (u *usecase) UpdateUserSchedule(ctx context.Context, userID int, s []du.UserSchedule, from string, to *string,oldFrom string, oldTo *string) error {




	exist, err := u.repo.WeeklyExist(ctx, userID, from, to,oldFrom, oldTo)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("Kolizja z innym tygodniem, (nie zaktualizowano)")
	}

	return u.repo.UpdateWeeklySchedule(ctx, userID, s, from, to,oldFrom, oldTo )
}


func (u *usecase) DeleteWeekly(ctx context.Context, userID int, from string, to *string) error {
	return u.repo.DeleteWeekly(ctx, userID, from, to)
}

// -------------------- OVERRIDES --------------------

func (u *usecase) ListOverrides(ctx context.Context, userID int, month string) ([]du.UserScheduleOverride, error) {
	return u.repo.ListOverrides(ctx, userID, month)
}

func (u *usecase) AddOverride(
	ctx context.Context,
	userID int,
	day string,
	planned int,
	earliestStart, latestStart, latestEnd *string,
	flex, spans bool,
) error {
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		return err
	}

	return u.repo.SetDayOverride(ctx, du.UserScheduleOverride{
		UserID:         userID,
		Day:            t,
		PlannedMinutes: planned,
		EarliestStart:  earliestStart,
		LatestStart:    latestStart,
		LatestEnd:      latestEnd,
		Flex:           flex,
		SpansMidnight:  spans,
	})
}

// -------------------- RAPORTY MIESIĘCZNE --------------------

func (u *usecase) GetMonthly(ctx context.Context, userID int, month string) (du.MonthlySchedule, error) {
	days, err := u.repo.GetMonthlySchedule(ctx, userID, month)
	if err != nil {
		return du.MonthlySchedule{}, err
	}

	return du.MonthlySchedule{
		UserID: userID,
		Month:  month,
		Days:   days,
	}, nil
}




func (u *usecase) GetMonthlyNew(ctx context.Context, userID int, month string) (du.MonthlyReport, error) {
	fromDate, toDate, err := monthBounds(month)
	if err != nil {
		return du.MonthlyReport{}, err
	}

	data, err := u.repo.GetMonthlyScheduleNew(ctx, userID, fromDate, toDate) // <- zwraca MonthlyReport
	if err != nil {
		return du.MonthlyReport{}, err
	}

	return data, nil
}


func monthBounds(month string) (time.Time, time.Time, error) {
	// zakładamy "YYYY-MM" np. "2025-11"
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid month format (expected YYYY-MM)")
	}

	// DATE w SQL nie potrzebuje strefy, ale trzymajmy spójnie UTC (albo użyj time.Local jeśli cały projekt tak działa)
	loc := time.UTC

	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	to := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, loc) // ostatni dzień miesiąca

	return from, to, nil
}



func (u *usecase) GetMonthlyForUsers(ctx context.Context, userIDs []int, month string) (du.MultiUserMonthlySchedule, error) {
	result := du.MultiUserMonthlySchedule{
		Month: month,
		Users: make([]du.CompactMonthlySchedule, 0, len(userIDs)),
	}

	for _, uid := range userIDs {
		m, err := u.GetMonthly(ctx, uid, month)
		if err != nil {
			return du.MultiUserMonthlySchedule{}, err
		}

		compact := du.CompactMonthlySchedule{
			UserID: m.UserID,
			Month:  m.Month,
			Days:   make([]du.CompactDailyRecord, len(m.Days)),
		}

		for i, d := range m.Days {
			compact.Days[i] = du.CompactDailyRecord{
				Date:           d.Date,
				Weekday:        d.Weekday,
				PlannedMinutes: d.PlannedMinutes,
				PlanSource:     d.PlanSource,
				LeaveMinutes:   d.LeaveMinutes,
				LeaveType:      d.LeaveType,
				WorkedMinutes:  d.WorkedMinutes,
				BreakMinutes:   d.BreakMinutes,
				Overtime:       d.Overtime,
			}
		}

		result.Users = append(result.Users, compact)
	}

	return result, nil
}

func (u *usecase) GetMonthlyForAllEmployees(
	ctx context.Context,
	month string,
	departmentIDs []int,
) (du.MultiUserMonthlySchedule, error) {

	ids, err := u.repo.ListEmployeeIDs(ctx, departmentIDs)
	if err != nil {
		return du.MultiUserMonthlySchedule{}, err
	}

	if len(ids) == 0 {
		return du.MultiUserMonthlySchedule{
			Month: month,
			Users: []du.CompactMonthlySchedule{},
		}, nil
	}

	return u.GetMonthlyForUsers(ctx, ids, month)
}

// -------------------- TIME POLICIES --------------------

func (u *usecase) GetActiveTimePolicy(ctx context.Context, day time.Time) (*du.TimePolicy, error) {
	return u.repo.GetActiveTimePolicy(ctx, day)
}

func (u *usecase) ListTimePolicies(ctx context.Context) ([]du.TimePolicy, error) {
	return u.repo.ListTimePolicies(ctx)
}

func (u *usecase) SaveTimePolicy(ctx context.Context, p du.TimePolicy) error {

exist, err := u.repo.TimePoliciesExist(ctx, p)
    if err != nil {
        return fmt.Errorf("błąd podczas sprawdzania kolizji: %w", err)
    }
    if exist {
        return fmt.Errorf("istnieje polityka czasu pracy w tym przedziale")
    }
	return u.repo.SaveTimePolicy(ctx, p)
}

func (u *usecase) GenerateMonthlyReportPDF(ctx context.Context, userID int, month string) (string, error) {
	if u.reportGen == nil {
		return "", errors.New("generator raportów PDF nie jest skonfigurowany")
	}
	report, err := u.GetMonthlyNew(ctx, userID, month)
	if err != nil {
		return "", fmt.Errorf("błąd pobierania raportu miesięcznego: %w", err)
	}
	return u.reportGen.GenerateMonthlyPDF(ctx, report, fmt.Sprintf("%d", userID), month)
}

