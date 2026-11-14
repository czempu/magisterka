package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	du "erpbadger/internal/domain/schedule"
)

func (r *ScheduleRepo) DeleteTimePolicies(ctx context.Context, id int) error {
	query := "delete from time_policies where id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ScheduleRepo) TimePoliciesExist(ctx context.Context, p du.TimePolicy) (bool, error) {
	if p.ID == 0 {
		const q = `
SELECT EXISTS (
  SELECT 1
  FROM time_policies e
  WHERE ? <= COALESCE(e.effective_to, '9999-12-31')
    AND COALESCE(?, '9999-12-31') >= e.effective_from
) AS is_collision;
`
		var exists bool
		if err := r.db.QueryRowContext(ctx, q, p.EffectiveFrom, p.EffectiveTo).Scan(&exists); err != nil {
			return false, err
		}
		return exists, nil
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM time_policies e
  WHERE ? <= COALESCE(e.effective_to, '9999-12-31')
    AND COALESCE(?, '9999-12-31') >= e.effective_from
    AND e.id <> ?
) AS is_collision;
`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, p.EffectiveFrom, p.EffectiveTo, p.ID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ScheduleRepo) GetActiveTimePolicy(ctx context.Context, day time.Time) (*du.TimePolicy, error) {
	const q = `
		SELECT
			id,
			effective_from,
			effective_to,
			earliest_start,
			latest_start,
			latest_end,
			flex
		FROM time_policies
		WHERE effective_from <= ?
		  AND (effective_to IS NULL OR effective_to >= ?)
		ORDER BY effective_from DESC
		LIMIT 1;
	`

	d := day.Format("2006-01-02")

	var (
		id            int64
		effFromStr    string
		effToStr      sql.NullString
		earliestStr   sql.NullString
		latestStartStr sql.NullString
		latestEndStr  sql.NullString
		flex          bool
	)

	err := r.db.QueryRowContext(ctx, q, d, d).Scan(
		&id,
		&effFromStr,
		&effToStr,
		&earliestStr,
		&latestStartStr,
		&latestEndStr,
		&flex,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // brak polityki na ten dzień
		}
		return nil, err
	}

	effFrom, err := time.Parse("2006-01-02", effFromStr)
	if err != nil {
		return nil, fmt.Errorf("parse effective_from (%s): %w", effFromStr, err)
	}

	var effTo *time.Time
	if effToStr.Valid && effToStr.String != "" {
		t, err := time.Parse("2006-01-02", effToStr.String)
		if err != nil {
			return nil, fmt.Errorf("parse effective_to (%s): %w", effToStr.String, err)
		}
		effTo = &t
	}

	tp := &du.TimePolicy{
		ID:            id,
		EffectiveFrom: effFrom,
		EffectiveTo:   effTo,
		Flex:          flex,
	}

	if earliestStr.Valid && earliestStr.String != "" {
		s := earliestStr.String
		if len(s) >= 5 {
			s = s[:5]
		}
		val := s
		tp.EarliestStart = &val
	}
	if latestStartStr.Valid && latestStartStr.String != "" {
		s := latestStartStr.String
		if len(s) >= 5 {
			s = s[:5]
		}
		val := s
		tp.LatestStart = &val
	}
	if latestEndStr.Valid && latestEndStr.String != "" {
		s := latestEndStr.String
		if len(s) >= 5 {
			s = s[:5]
		}
		val := s
		tp.LatestEnd = &val
	}

	return tp, nil
}

func (r *ScheduleRepo) ListTimePolicies(ctx context.Context) ([]du.TimePolicy, error) {
	const q = `
		SELECT
			id,
			effective_from,
			effective_to,
			earliest_start,
			latest_start,
			latest_end,
			flex
		FROM time_policies
		ORDER BY effective_from;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []du.TimePolicy

	for rows.Next() {
		var (
			id            int64
			effFromStr    string
			effToStr      sql.NullString
			earliestStr   sql.NullString
			latestStartStr sql.NullString
			latestEndStr  sql.NullString
			flex          bool
		)

		if err := rows.Scan(
			&id,
			&effFromStr,
			&effToStr,
			&earliestStr,
			&latestStartStr,
			&latestEndStr,
			&flex,
		); err != nil {
			return nil, err
		}

		effFrom, err := time.Parse("2006-01-02", effFromStr)
		if err != nil {
			return nil, fmt.Errorf("parse effective_from (%s): %w", effFromStr, err)
		}

		var effTo *time.Time
		if effToStr.Valid && effToStr.String != "" {
			t, err := time.Parse("2006-01-02", effToStr.String)
			if err != nil {
				return nil, fmt.Errorf("parse effective_to (%s): %w", effToStr.String, err)
			}
			effTo = &t
		}

		tp := du.TimePolicy{
			ID:            id,
			EffectiveFrom: effFrom,
			EffectiveTo:   effTo,
			Flex:          flex,
		}

		if earliestStr.Valid && earliestStr.String != "" {
			s := earliestStr.String
			if len(s) >= 5 {
				s = s[:5]
			}
			val := s
			tp.EarliestStart = &val
		}
		if latestStartStr.Valid && latestStartStr.String != "" {
			s := latestStartStr.String
			if len(s) >= 5 {
				s = s[:5]
			}
			val := s
			tp.LatestStart = &val
		}
		if latestEndStr.Valid && latestEndStr.String != "" {
			s := latestEndStr.String
			if len(s) >= 5 {
				s = s[:5]
			}
			val := s
			tp.LatestEnd = &val
		}

		result = append(result, tp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ScheduleRepo) SaveTimePolicy(ctx context.Context, p du.TimePolicy) error {
	effFrom := p.EffectiveFrom.Format("2006-01-02")
	var effTo any
	if p.EffectiveTo != nil {
		effTo = p.EffectiveTo.Format("2006-01-02")
	} else {
		effTo = nil
	}

	var earliest any
	if p.EarliestStart != nil && *p.EarliestStart != "" {
		earliest = *p.EarliestStart
	} else {
		earliest = nil
	}

	var latestStart any
	if p.LatestStart != nil && *p.LatestStart != "" {
		latestStart = *p.LatestStart
	} else {
		latestStart = nil
	}

	var latestEnd any
	if p.LatestEnd != nil && *p.LatestEnd != "" {
		latestEnd = *p.LatestEnd
	} else {
		latestEnd = nil
	}

	if p.ID == 0 {
		const q = `
			INSERT INTO time_policies (
				effective_from,
				effective_to,
				earliest_start,
				latest_start,
				latest_end,
				flex
			) VALUES (?, ?, ?, ?, ?, ?);
		`

		res, err := r.db.ExecContext(ctx, q,
			effFrom,
			effTo,
			earliest,
			latestStart,
			latestEnd,
			p.Flex,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err == nil {
			p.ID = id
		}
		return nil
	}

	const q = `
		UPDATE time_policies
		SET
			effective_from = ?,
			effective_to   = ?,
			earliest_start = ?,
			latest_start   = ?,
			latest_end     = ?,
			flex           = ?
		WHERE id = ?;
	`

	_, err := r.db.ExecContext(ctx, q,
		effFrom,
		effTo,
		earliest,
		latestStart,
		latestEnd,
		p.Flex,
		p.ID,
	)
	return err
}
