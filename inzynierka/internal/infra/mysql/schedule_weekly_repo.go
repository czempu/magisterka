package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	du "erpbadger/internal/domain/schedule"
)

// CheckWeekSchedule:
// 1) sprawdza, czy istnieje już grafik tygodniowy dla usera w danym przedziale
// 2) jeśli nie – zakłada pusty grafik (dow 1–7, planned_minutes = 0) dla wskazanego zakresu
func (r *ScheduleRepo) CheckWeekSchedule(ctx context.Context, userID int, from, to *string) (int, error) {
	if from == nil || *from == "" {
		return 0, fmt.Errorf("musi być określony początek obowiązywania grafiku (effective_from)")
	}

	fromVal := *from

	var toVal any
	if to != nil && *to != "" {
		toVal = *to
	} else {
		toVal = nil
	}

	const checkQuery = `
		SELECT 1 
		FROM user_schedules_weekly usw
		WHERE usw.user_id = ?
		  AND usw.effective_from <= ?
		  AND (usw.effective_to >= ? OR usw.effective_to IS NULL)
		LIMIT 1;
	`

	var exists int
	err := r.db.QueryRowContext(ctx, checkQuery, userID, fromVal, toVal).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("błąd podczas sprawdzania istniejącego grafiku: %w", err)
	}

	if err == nil && exists == 1 {
		return 0, fmt.Errorf("dla użytkownika %d istnieje już tygodniowy grafik w tym przedziale dat", userID)
	}

	const insertQuery = `
		INSERT INTO user_schedules_weekly (
			user_id,
			dow,
			planned_minutes,
			earliest_start,
			latest_start,
			latest_end,
			spans_midnight,
			flex,
			effective_from,
			effective_to
		)
		VALUES (?, ?, 0, NULL, NULL, NULL, 0, 1, ?, ?);
	`

	for dow := 1; dow <= 7; dow++ {
		if _, err := r.db.ExecContext(ctx, insertQuery, userID, dow, fromVal, toVal); err != nil {
			return 0, fmt.Errorf("błąd podczas dodawania tygodniowego grafiku (dow=%d): %w", dow, err)
		}
	}

	return 1, nil
}

func (r *ScheduleRepo) GetWeeklySchedule(ctx context.Context, userID int, from, to *string) ([]du.UserSchedule, error) {
	query := `
		SELECT 
			id,
			user_id,
			dow,
			planned_minutes,
			earliest_start,
			latest_start,
			latest_end,
			spans_midnight,
			flex,
			effective_from,
			effective_to
		FROM user_schedules_weekly
		WHERE user_id = ?
	`
	args := []any{userID}

	if from == nil {
		query += " AND effective_from IS NULL"
	} else {
		query += " AND effective_from = ?"
		args = append(args, *from)
	}

	if to == nil {
		query += " AND effective_to IS NULL"
	} else {
		query += " AND effective_to = ?"
		args = append(args, *to)
	}

	query += " ORDER BY dow"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []du.UserSchedule
	for rows.Next() {
		var s du.UserSchedule

		var effFromStr, effToStr sql.NullString
		var earliestStr, latestStartStr, latestEndStr sql.NullString
		var spans, flex sql.NullBool

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.DayOfWeek,
			&s.PlannedMinutes,
			&earliestStr,
			&latestStartStr,
			&latestEndStr,
			&spans,
			&flex,
			&effFromStr,
			&effToStr,
		); err != nil {
			return nil, err
		}

		if effFromStr.Valid {
			tmp := effFromStr.String
			s.EffectiveFrom = &tmp
		}
		if effToStr.Valid {
			tmp := effToStr.String
			s.EffectiveTo = &tmp
		}

		if earliestStr.Valid {
			tmp := earliestStr.String[:5]
			s.EarliestStart = &tmp
		}
		if latestStartStr.Valid {
			tmp := latestStartStr.String[:5]
			s.LatestStart = &tmp
		}
		if latestEndStr.Valid {
			tmp := latestEndStr.String[:5]
			s.LatestEnd = &tmp
		}

		if spans.Valid {
			s.SpansMidnight = spans.Bool
		}
		if flex.Valid {
			s.Flex = flex.Bool
		} else {
			s.Flex = true
		}

		res = append(res, s)
	}

	return res, rows.Err()
}

func (r *ScheduleRepo) GetWeeks(ctx context.Context, userID int) ([]du.UserScheduleWeeks, error) {
	q := `
		SELECT user_id, effective_from, effective_to
		FROM user_schedules_weekly
		WHERE user_id = ?
		GROUP BY user_id, effective_from, effective_to;
	`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []du.UserScheduleWeeks
	for rows.Next() {
		var o du.UserScheduleWeeks
		if err := rows.Scan(&o.ID, &o.EffectiveFrom, &o.EffectiveTo); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func (r *ScheduleRepo) DeleteWeekly(ctx context.Context, userID int, from string, to *string) error {
	query := "DELETE FROM user_schedules_weekly WHERE user_id = ? AND effective_from = ?"
	args := []interface{}{userID, from}

	if to == nil {
		query += " AND effective_to IS NULL"
	} else {
		query += " AND effective_to = ?"
		args = append(args, *to)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *ScheduleRepo) WeeklyExist(ctx context.Context, userID int, from string, to *string, oldFrom string, oldTo *string) (bool, error) {
	q := "WITH tmp AS (SELECT id FROM user_schedules_weekly WHERE user_id = ? "
	args := []interface{}{userID}

	if to != nil {
		q += "AND effective_to = ? "
		args = append(args, *to)
	} else {
		q += "AND effective_to IS NULL "
	}

	q += "AND effective_from = ?) "
	args = append(args, from)

	q += `SELECT EXISTS (
	  SELECT 1
	  FROM user_schedules_weekly e
	  WHERE e.user_id = ?
		-- [from] <= koniec
		AND ? <= COALESCE(e.effective_to, '9999-12-31')
		-- [to] >= start
		AND ? >= e.effective_from
		AND e.id NOT IN (SELECT id FROM tmp)
	) AS is_collision;`

	args = append(args, userID)
	args = append(args, oldFrom)

	if oldTo != nil {
		args = append(args, *oldTo)
	} else {
		args = append(args, "9999-12-31")
	}

	var exists bool
	err := r.db.QueryRowContext(ctx, q, args...).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ScheduleRepo) UpdateWeeklySchedule(ctx context.Context, userID int, schedule []du.UserSchedule, from string, to *string, oldFrom string, oldTo *string) error {
	whereParts := []string{"user_id = ?"}
	args := []any{userID}

	whereParts = append(whereParts, "effective_from = ?")
	args = append(args, from)

	if to == nil {
		whereParts = append(whereParts, "effective_to IS NULL")
	} else {
		whereParts = append(whereParts, "effective_to = ?")
		args = append(args, *to)
	}

	whereClause := " WHERE " + strings.Join(whereParts, " AND ")

	baseQuery := `
		UPDATE user_schedules_weekly
		SET 
			planned_minutes = ?,
			earliest_start  = ?,
			latest_start    = ?,
			latest_end      = ?,
			spans_midnight  = ?,
			flex            = ?,
			effective_from  = ?,
			effective_to    = ?
	` + whereClause + " AND dow = ?"

	toTimeVal := func(s *string) any {
		if s == nil || *s == "" {
			return nil
		}
		return *s + ":00"
	}

	for _, s := range schedule {
		spansVal := 0
		if s.SpansMidnight {
			spansVal = 1
		}
		flexVal := 0
		if s.Flex {
			flexVal = 1
		}

		allArgs := []any{
			s.PlannedMinutes,
			toTimeVal(s.EarliestStart),
			toTimeVal(s.LatestStart),
			toTimeVal(s.LatestEnd),
			spansVal,
			flexVal,
			s.EffectiveFrom,
			s.EffectiveTo,
		}
		allArgs = append(allArgs, args...)
		allArgs = append(allArgs, s.DayOfWeek)

		if _, err := r.db.ExecContext(ctx, baseQuery, allArgs...); err != nil {
			return err
		}
	}

	return nil
}
