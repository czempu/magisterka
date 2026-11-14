package mysql

import (
	"context"
	"database/sql"
	"time"

	du "erpbadger/internal/domain/schedule"
)

func (r *ScheduleRepo) DeleteOverride(ctx context.Context, id int) error {
	query := "delete from user_schedule_overrides where id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ScheduleRepo) GetDayOverride(ctx context.Context, userID int, day string) (*du.UserScheduleOverride, error) {
	q := `
		SELECT 
			id,
			user_id,
			day,
			planned_minutes,
			earliest_start,
			latest_start,
			latest_end,
			spans_midnight,
			flex
		FROM user_schedule_overrides 
		WHERE user_id = ? AND day = ?;
	`

	var o du.UserScheduleOverride
	var dayStr string
	var earliestStr, latestStartStr, latestEndStr sql.NullString
	var spans, flex sql.NullBool

	err := r.db.QueryRowContext(ctx, q, userID, day).Scan(
		&o.ID,
		&o.UserID,
		&dayStr,
		&o.PlannedMinutes,
		&earliestStr,
		&latestStartStr,
		&latestEndStr,
		&spans,
		&flex,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	t, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		o.Day = time.Time{}
	} else {
		o.Day = t
	}

	if earliestStr.Valid {
		tmp := earliestStr.String[:5]
		o.EarliestStart = &tmp
	}
	if latestStartStr.Valid {
		tmp := latestStartStr.String[:5]
		o.LatestStart = &tmp
	}
	if latestEndStr.Valid {
		tmp := latestEndStr.String[:5]
		o.LatestEnd = &tmp
	}

	if spans.Valid {
		o.SpansMidnight = spans.Bool
	}
	if flex.Valid {
		o.Flex = flex.Bool
	} else {
		o.Flex = true
	}

	return &o, nil
}

func (r *ScheduleRepo) SetDayOverride(ctx context.Context, override du.UserScheduleOverride) error {
	toTimeVal := func(s *string) any {
		if s == nil || *s == "" {
			return nil
		}
		return *s + ":00"
	}

	spansVal := 0
	if override.SpansMidnight {
		spansVal = 1
	}
	flexVal := 0
	if override.Flex {
		flexVal = 1
	}

	q := `
		INSERT INTO user_schedule_overrides (
			user_id,
			day,
			planned_minutes,
			earliest_start,
			latest_start,
			latest_end,
			spans_midnight,
			flex
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			planned_minutes = VALUES(planned_minutes),
			earliest_start  = VALUES(earliest_start),
			latest_start    = VALUES(latest_start),
			latest_end      = VALUES(latest_end),
			spans_midnight  = VALUES(spans_midnight),
			flex            = VALUES(flex);
	`

	_, err := r.db.ExecContext(
		ctx,
		q,
		override.UserID,
		override.Day,
		override.PlannedMinutes,
		toTimeVal(override.EarliestStart),
		toTimeVal(override.LatestStart),
		toTimeVal(override.LatestEnd),
		spansVal,
		flexVal,
	)
	return err
}

func (r *ScheduleRepo) ListOverrides(ctx context.Context, userID int, month string) ([]du.UserScheduleOverride, error) {
	q := `
		SELECT 
			id,
			user_id,
			day,
			planned_minutes,
			earliest_start,
			latest_start,
			latest_end,
			spans_midnight,
			flex
		FROM user_schedule_overrides
		WHERE user_id = ? 
		  AND LEFT(day, 7) = ? 
		ORDER BY day;
	`

	rows, err := r.db.QueryContext(ctx, q, userID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []du.UserScheduleOverride
	for rows.Next() {
		var o du.UserScheduleOverride
		var dayStr string
		var earliestStr, latestStartStr, latestEndStr sql.NullString
		var spans, flex sql.NullBool

		if err := rows.Scan(
			&o.ID,
			&o.UserID,
			&dayStr,
			&o.PlannedMinutes,
			&earliestStr,
			&latestStartStr,
			&latestEndStr,
			&spans,
			&flex,
		); err != nil {
			return nil, err
		}

		t, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			o.Day = time.Time{}
		} else {
			o.Day = t
		}

		if earliestStr.Valid {
			tmp := earliestStr.String[:5]
			o.EarliestStart = &tmp
		}
		if latestStartStr.Valid {
			tmp := latestStartStr.String[:5]
			o.LatestStart = &tmp
		}
		if latestEndStr.Valid {
			tmp := latestEndStr.String[:5]
			o.LatestEnd = &tmp
		}

		if spans.Valid {
			o.SpansMidnight = spans.Bool
		}
		if flex.Valid {
			o.Flex = flex.Bool
		} else {
			o.Flex = true
		}

		result = append(result, o)
	}
	return result, rows.Err()
}
