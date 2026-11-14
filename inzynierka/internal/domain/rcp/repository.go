package rcp

import "context"
import "time"

type Repository interface {
 GetMeWorkStatus(ctx context.Context, userID int) ([]WorkStatus, error)
 PostMeWorkStart(ctx context.Context, userID int) ([]WorkStart, error)
 PostMeWorkStop(ctx context.Context, userID int) ([]WorkStop, error)
GetMeWorkTimesheet(ctx context.Context, userID int) ([]WorkTimesheet, error)
	ListActiveEmployees(day time.Time) ([]int, error)
	GetPlannedMinutes(userID int, day time.Time) (int, error)
	GetWorkedAndBreakMinutes(userID int, day time.Time) (workedMinutes int, breakMinutes int, err error)
	UpsertDailyTotal(dt DailyTotal) error
}
