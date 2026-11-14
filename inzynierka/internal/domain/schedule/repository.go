package schedule

import (
	"context"
	"time"
)

type Repository interface {
	// -------------------- GRAFIK TYGODNIOWY --------------------
	GetWeeklySchedule(ctx context.Context, userID int, from, to *string) ([]UserSchedule, error)
	GetWeeks(ctx context.Context, userID int) ([]UserScheduleWeeks, error)
	UpdateWeeklySchedule(ctx context.Context, userID int, schedule []UserSchedule, from string, to *string,oldFrom string, oldTo *string) error
	CheckWeekSchedule(ctx context.Context, userID int, from, to *string) (int, error)

	// -------------------- OVERRIDES --------------------
	GetDayOverride(ctx context.Context, userID int, day string) (*UserScheduleOverride, error)
	SetDayOverride(ctx context.Context, override UserScheduleOverride) error
	ListOverrides(ctx context.Context, userID int, month string) ([]UserScheduleOverride, error)

	// -------------------- RAPORT MIESIĘCZNY --------------------
	GetMonthlySchedule(ctx context.Context, userID int, month string) ([]DailyRecord, error)
GetMonthlyScheduleNew(ctx context.Context, userID int, fromDate, toDate time.Time) (MonthlyReport, error)
	// -------------------- PRACOWNICY (DO RAPORTÓW ZBIORCZYCH) --------------------
	ListEmployeeIDs(ctx context.Context, departmentIDs []int) ([]int, error)

	// -------------------- TIME POLICIES --------------------
	GetActiveTimePolicy(ctx context.Context, day time.Time) (*TimePolicy, error)
	ListTimePolicies(ctx context.Context) ([]TimePolicy, error)
	SaveTimePolicy(ctx context.Context, p TimePolicy) error



	DeleteTimePolicies(ctx context.Context, id int) error
	TimePoliciesExist(ctx context.Context, p TimePolicy) (bool,error)
WeeklyExist(ctx context.Context, UserID int, from string, to *string,oldFrom string, oldTo *string) (bool,error)



	DeleteOverride(ctx context.Context, id int) error
	DeleteWeekly(ctx context.Context, UserID int, from string, to *string) error
}

