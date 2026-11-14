package mysql

import (
	"database/sql"

	du "erpbadger/internal/domain/schedule"
)

// ScheduleRepo implements du.Repository for schedule management.
// Methods are modularized into:
// - schedule_weekly_repo.go: Weekly recurring schedules
// - schedule_override_repo.go: Specific day overrides
// - time_policy_repo.go: Organization time policies
// - schedule_monthly_repo.go: Monthly schedule reports and KPI aggregations
type ScheduleRepo struct {
	db *sql.DB
}

// NewScheduleRepo creates a new instance of du.Repository backed by MySQL.
func NewScheduleRepo(db *sql.DB) du.Repository {
	return &ScheduleRepo{db: db}
}
