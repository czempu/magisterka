package mysql

import (
    "context"
    "database/sql"
    "erpbadger/internal/domain/rcp"
    "time"
    
)

type RcpRepo struct {
    db *sql.DB
}

func NewRcpRepo(db *sql.DB) rcp.Repository {
    return &RcpRepo{db: db}
}


func (r *RcpRepo) GetMeWorkStatus(ctx context.Context,userID int) ([]rcp.WorkStatus, error) {
   
	q := `
SELECT 
  (open_id IS NOT NULL) AS is_active,
COALESCE(open_started_at, '') AS open_started_at
FROM (
  SELECT 
    (SELECT id        FROM timesheet_entries 
     WHERE user_id = ? AND ended_at IS NULL 
     ORDER BY started_at DESC LIMIT 1) AS open_id,
    (SELECT started_at FROM timesheet_entries 
     WHERE user_id = ? AND ended_at IS NULL 
     ORDER BY started_at DESC LIMIT 1) AS open_started_at
) t;

	`
    rows, err := r.db.QueryContext(ctx, q, userID, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var modules []rcp.WorkStatus
    for rows.Next() {
        var mi rcp.WorkStatus
        if err := rows.Scan(&mi.Status, &mi.Data); err != nil {
            return nil, err
        }
        modules = append(modules, mi)
    }
    return modules, rows.Err()
}

func (r *RcpRepo) PostMeWorkStart(ctx context.Context,userID int) ([]rcp.WorkStart, error) {
   
  query := `
INSERT INTO timesheet_entries (user_id, started_at, type, note)
SELECT ?, NOW(), 'work', 'start zmiany'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 
  FROM timesheet_entries 
  WHERE user_id = ? AND ended_at IS NULL
);
  `
    _, err := r.db.ExecContext(ctx, query, userID, userID)
    if err != nil {   

    var modules []rcp.WorkStart
        var mi rcp.WorkStart
	mi.Status = "Błąd"
	mi.Data = "Błąd"

        modules = append(modules, mi)
    return modules, nil
    }

    var modules []rcp.WorkStart
        var mi rcp.WorkStart
	mi.Status = "Rozpoczęto prace"
	mi.Data = "Now"

        modules = append(modules, mi)
    return modules, nil
}

func (r *RcpRepo) PostMeWorkStop(ctx context.Context,userID int) ([]rcp.WorkStop, error) {
   
  query := `
UPDATE timesheet_entries
SET ended_at = NOW(), note = 'koniec zmiany'
WHERE user_id = ? AND ended_at IS NULL
ORDER BY started_at DESC
LIMIT 1;
  `
    _, err := r.db.ExecContext(ctx, query, userID)
    if err != nil {   
     var modules []rcp.WorkStop
        var mi rcp.WorkStop
	mi.Status = "Błąd"

        modules = append(modules, mi)
    return modules, nil
    }

    var modules []rcp.WorkStop
        var mi rcp.WorkStop
 
	mi.Status = "Zakończono pracę"



	modules = append(modules, mi)
    return modules, nil
}


func (r *RcpRepo) GetMeWorkTimesheet(ctx context.Context, UserID int) ([]rcp.WorkTimesheet, error) {
   
	q := `
SELECT id, user_id, started_at, COALESCE(ended_at,NOW()) AS ended_at, TIMESTAMPDIFF(MINUTE, started_at, COALESCE(ended_at,NOW())) AS minutes
FROM timesheet_entries
WHERE user_id = ?
ORDER BY started_at DESC; `
    rows, err := r.db.QueryContext(ctx, q, UserID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var modules []rcp.WorkTimesheet
    for rows.Next() {
        var mi rcp.WorkTimesheet
        if err := rows.Scan(&mi.ID, &mi.UserID, &mi.Startedat, &mi.Endedat,&mi.Minutes); err != nil {
            return nil, err
        }
        modules = append(modules, mi)
    }
    return modules, rows.Err()
}




func (r *RcpRepo) ListActiveEmployees(day time.Time) ([]int, error) {
	const q = `
		SELECT id
		FROM RCP.users
		WHERE is_employee = 1;
	`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}






func (r *RcpRepo) GetPlannedMinutes(userID int, day time.Time) (int, error) {
	dayStr := day.Format("2006-01-02")

	baseMinutes, err := r.getBaseMinutes(userID, dayStr)
	if err != nil {
		return 0, err
	}
	planned := baseMinutes

	isHoliday, err := r.isHoliday(userID, dayStr)
	if err != nil {
		return 0, err
	}
	if isHoliday {
		planned = 0
	}

	if m, ok, err := r.getWeeklyScheduleMinutes(userID, day); err != nil {
		return 0, err
	} else if ok {
		planned = m
	}

	if m, ok, err := r.getOverrideMinutes(userID, dayStr); err != nil {
		return 0, err
	} else if ok {
		planned = m
	}

	if planned < 0 {
		planned = 0
	}

	return planned, nil
}

func (r *RcpRepo) getBaseMinutes(userID int, day string) (int, error) {
	const q = `
SELECT 
  COALESCE(
    (
      SELECT et.nominal_day_minutes
      FROM RCP.employment_terms et
      WHERE et.user_id = u.id
        AND ? BETWEEN et.effective_from AND COALESCE(et.effective_to, '9999-12-31')
      ORDER BY et.effective_from DESC
      LIMIT 1
    ),
    u.nominal_day_minutes,
    0
  ) AS minutes
FROM RCP.users u
WHERE u.id = ?
`
	var minutes sql.NullInt64
	if err := r.db.QueryRow(q, day, userID).Scan(&minutes); err != nil {
		return 0, err
	}
	if !minutes.Valid {
		return 0, nil
	}
	return int(minutes.Int64), nil
}

func (r *RcpRepo) isHoliday(userID int, day string) (bool, error) {
	const q = `
SELECT 1
FROM RCP.employee_assignments ea
JOIN RCP.calendar_holidays ch
  ON ch.work_calendar_id = ea.work_calendar_id
WHERE ea.employee_id = ?
  AND ch.day = ?
  AND ? BETWEEN ea.effective_from AND COALESCE(ea.effective_to, '9999-12-31')
LIMIT 1;
`

	var dummy int
	err := r.db.QueryRow(q, userID, day, day).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *RcpRepo) getWeeklyScheduleMinutes(userID int, day time.Time) (int, bool, error) {
	dow := int(day.Weekday()) // upewnij się, że mapping = DOW w bazie

	const q = `
SELECT usw.planned_minutes
FROM RCP.user_schedules_weekly usw
WHERE usw.user_id = ?
  AND ? BETWEEN usw.effective_from AND COALESCE(usw.effective_to, '9999-12-31')
  AND usw.dow = ?
ORDER BY usw.effective_from DESC
LIMIT 1;
`
	var minutes sql.NullInt64
	err := r.db.QueryRow(q, userID, day.Format("2006-01-02"), dow).Scan(&minutes)
	if err == sql.ErrNoRows || !minutes.Valid {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(minutes.Int64), true, nil
}

func (r *RcpRepo) getOverrideMinutes(userID int, day string) (int, bool, error) {
	const q = `
SELECT uso.planned_minutes
FROM RCP.user_schedule_overrides uso
WHERE uso.user_id = ? AND uso.day = ?
LIMIT 1;
`
	var minutes sql.NullInt64
	err := r.db.QueryRow(q, userID, day).Scan(&minutes)
	if err == sql.ErrNoRows || !minutes.Valid {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(minutes.Int64), true, nil
}





func (r *RcpRepo) GetWorkedAndBreakMinutes(userID int, day time.Time) (int, int, error) {
	start := day.Format("2006-01-02")

	const q = `
SELECT
  COALESCE(SUM(CASE WHEN te.type = 'work' THEN TIMESTAMPDIFF(MINUTE, te.started_at, te.ended_at) ELSE 0 END), 0) AS work_minutes,
  COALESCE(SUM(CASE WHEN te.type = 'break' THEN TIMESTAMPDIFF(MINUTE, te.started_at, te.ended_at) ELSE 0 END), 0) AS break_minutes
FROM RCP.timesheet_entries te
WHERE te.user_id = ?
  AND te.started_at >= ?
  AND te.started_at < DATE_ADD(?, INTERVAL 1 DAY);
`
	var work, brk int
	if err := r.db.QueryRow(q, userID, start, start).Scan(&work, &brk); err != nil {
		return 0, 0, err
	}
	if work < 0 {
		work = 0
	}
	if brk < 0 {
		brk = 0
	}
	return work, brk, nil
}

func (r *RcpRepo) UpsertDailyTotal(dt rcp.DailyTotal) error {
	const q = `
INSERT INTO RCP.timesheet_daily_totals
  (user_id, day, worked_minutes, break_minutes, overtime_minutes, absence_minutes)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  worked_minutes   = VALUES(worked_minutes),
  break_minutes    = VALUES(break_minutes),
  overtime_minutes = VALUES(overtime_minutes),
  absence_minutes  = VALUES(absence_minutes);
`
	dayStr := dt.Day.Format("2006-01-02")

	_, err := r.db.Exec(q,
		dt.UserID,
		dayStr,
		dt.WorkedMinutes,
		dt.BreakMinutes,
		dt.OvertimeMinutes,
		dt.AbsenceMinutes,
	)
	return err
}
