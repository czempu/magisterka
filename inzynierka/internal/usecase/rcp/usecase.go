package rcp

import (
	"context"
	"fmt"
	"time"

	du "erpbadger/internal/domain/rcp"
	"erpbadger/internal/domain/schedule"
)

var p = fmt.Print



// Usecase – interfejs dla handlera/middleware
type Usecase interface {
 GetMeWorkStatus(ctx context.Context, userID int) ([]du.WorkStatus, error)
 PostMeWorkStart(ctx context.Context, userID int) ([]du.WorkStart, error)
 PostMeWorkStop(ctx context.Context, userID int) ([]du.WorkStop, error)
RecalculateDailyTotalsForDate(ctx context.Context, day time.Time) (error)
GetMeWorkTimesheet(ctx context.Context, userID int) ([]du.WorkTimesheet, error)

}

type usecase struct {
    repo du.Repository
}

// NewUsecase – wstrzykujesz tutaj repo (fake lub SQL)
func NewUsecase(r du.Repository) Usecase {
    return &usecase{repo: r}
}


func (u *usecase) GetMeWorkStatus(ctx context.Context, userID int) ([]du.WorkStatus, error) {
    return u.repo.GetMeWorkStatus(ctx, userID)
}

func (u *usecase) PostMeWorkStart(ctx context.Context, userID int) ([]du.WorkStart, error) {
    return u.repo.PostMeWorkStart(ctx, userID)
}

func (u *usecase) PostMeWorkStop(ctx context.Context, userID int) ([]du.WorkStop, error) {
    return u.repo.PostMeWorkStop(ctx, userID)
}

func (u *usecase) GetMeWorkTimesheet(ctx context.Context, userID int) ([]du.WorkTimesheet, error) {
    return u.repo.GetMeWorkTimesheet(ctx, userID)
}






// RecalculateDailyTotalsForDate – liczy wszystkie osoby dla jednego dnia.
func (u *usecase) RecalculateDailyTotalsForDate(ctx context.Context, day time.Time) error {
	day = truncateToDate(day)

	userIDs, err := u.repo.ListActiveEmployees(day)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := u.RecalculateDailyTotalForUserAndDate(ctx, userID, day); err != nil {
			// Tu możesz zamiast return dodać logowanie i lecieć dalej,
			// żeby jeden user nie wywalił całego crona.
			return err
		}
	}

	return nil
}

// RecalculateDailyTotalForUserAndDate – przydatne do testów i ręcznego przeliczenia.
func (u *usecase) RecalculateDailyTotalForUserAndDate(
	ctx context.Context,
	userID int,
	day time.Time,
) error {
	day = truncateToDate(day)

	planned, err := u.repo.GetPlannedMinutes(userID, day)
	if err != nil {
		return err
	}

	workedRaw, breakMinutes, err := u.repo.GetWorkedAndBreakMinutes(userID, day)
	if err != nil {
		return err
	}

	calc := schedule.CalculateDailyWorkTime(planned, workedRaw)

	dt := du.DailyTotal{
		UserID:          userID,
		Day:             day,
		WorkedMinutes:   calc.WorkedMinutes,
		BreakMinutes:    breakMinutes,
		OvertimeMinutes: calc.OvertimeMinutes,
		AbsenceMinutes:  calc.AbsenceMinutes,
	}

	return u.repo.UpsertDailyTotal(dt)
}

func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}


















