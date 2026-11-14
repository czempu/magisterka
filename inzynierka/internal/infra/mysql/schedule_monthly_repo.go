package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	du "erpbadger/internal/domain/schedule"
)

func (r *ScheduleRepo) GetMonthlySchedule(ctx context.Context, userID int, month string) ([]du.DailyRecord, error) {
	fmt.Println("generuje nowy grafik dla", userID, month)

	rows, err := r.db.QueryContext(ctx, "CALL sp_get_user_monthly_schedule(?, ?)", userID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []du.DailyRecord

	for rows.Next() {
		var rec du.DailyRecord
		var entriesJSON []byte
		var leaveType sql.NullString

		if err := rows.Scan(
			&rec.Date,
			&rec.Weekday,
			&rec.PlannedMinutes,
			&rec.PlanSource,
			&rec.LeaveMinutes,
			&leaveType,
			&rec.WorkedMinutes,
			&rec.BreakMinutes,
			&rec.Overtime,
			&entriesJSON,
		); err != nil {
			return nil, err
		}

		if leaveType.Valid {
			rec.LeaveType = leaveType.String
		}

		if entriesJSON != nil {
			_ = json.Unmarshal(entriesJSON, &rec.Entries)
		}

		records = append(records, rec)
	}

	return records, rows.Err()
}

// -------------------- PRACOWNICY --------------------

func (r *ScheduleRepo) ListEmployeeIDs(ctx context.Context, departmentIDs []int) ([]int, error) {
	// Na razie: wszyscy pracownicy z users.is_employee = 1
	const q = `SELECT id FROM users WHERE is_employee = 1`

	rows, err := r.db.QueryContext(ctx, q)
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



func (r *ScheduleRepo) GetMonthlyScheduleNew(ctx context.Context, userID int, fromDate, toDate time.Time) (du.MonthlyReport, error) {
	if userID <= 0 {
		return du.MonthlyReport{}, errors.New("invalid userID")
	}
	if fromDate.After(toDate) {
		return du.MonthlyReport{}, errors.New("fromDate after toDate")
	}

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return du.MonthlyReport{}, err
	}
	defer conn.Close()

	// żeby nie ucinało JSON w GROUP_CONCAT
	if _, err := conn.ExecContext(ctx, `SET SESSION group_concat_max_len = 1048576`); err != nil {
		return du.MonthlyReport{}, err
	}

	// user-defined variables są per-połączenie
	if _, err := conn.ExecContext(ctx, `SET @p_user_id := ?`, userID); err != nil {
		return du.MonthlyReport{}, err
	}
	if _, err := conn.ExecContext(ctx, `SET @p_from_date := DATE(?)`, fromDate); err != nil {
		return du.MonthlyReport{}, err
	}
	if _, err := conn.ExecContext(ctx, `SET @p_to_date := DATE(?)`, toDate); err != nil {
		return du.MonthlyReport{}, err
	}
	if _, err := conn.ExecContext(ctx, `SET @p_year := YEAR(@p_from_date)`); err != nil {
		return du.MonthlyReport{}, err
	}
	if _, err := conn.ExecContext(ctx, `SET @p_month := MONTH(@p_from_date)`); err != nil {
		return du.MonthlyReport{}, err
	}

	const q = `


	 
	 
WITH
-- ============================================================
-- KALENDARZ DNI: bez RECURSIVE, na seq_0_to_31 (0..31)
-- ============================================================
calendar AS (
    SELECT DATE_ADD(@p_from_date, INTERVAL s.seq DAY) AS day
    FROM seq_0_to_31 s
    WHERE s.seq BETWEEN 0 AND DATEDIFF(@p_to_date, @p_from_date)
),

-- ============================================================
-- 0) EMPLOYMENT TERMS per dzień (FTE + nominal_day_minutes)
-- ============================================================
et_day AS (
    SELECT
        c.day,
        @p_user_id AS user_id,
        COALESCE(et.fte_percent, 1.00) AS fte_percent
    FROM calendar c
    LEFT JOIN LATERAL (
        SELECT fte_percent
        FROM employment_terms
        WHERE user_id = @p_user_id
          AND c.day >= effective_from
          AND (effective_to IS NULL OR c.day <= effective_to)
        ORDER BY effective_from DESC, id DESC
        LIMIT 1
    ) et ON TRUE
),

-- ============================================================
-- 0b) NOMINAŁ MIESIĘCZNY (Nominals)
--     Nominals ma teraz normalne wartości: hours, minutes
-- ============================================================
nom_month AS (
    SELECT
        @p_year  AS year,
        @p_month AS month,
        COALESCE(
            CAST(n.minutes AS DECIMAL(18,2)),
            CAST(n.hours   AS DECIMAL(18,2)) * 60
        ) AS nominal_minutes_100
    FROM Nominals n
    WHERE n.year = @p_year AND n.month = @p_month
),

month_meta AS (
    SELECT (DATEDIFF(@p_to_date, @p_from_date) + 1) AS days_in_scope
),

-- ============================================================
-- 1) KONTEKST DNIA: kalendarz + time policy
--     UWAGA: user_time_policies / rounding_policies usunięte -> grace_minutes = 0
-- ============================================================
ctx AS (
    SELECT
        c.day,
        @p_user_id AS user_id,

        ea.work_calendar_id,

        -- opcjonalnie: identyfikator polityki wybranej dla danego dnia (przydaje się do debugowania)
        tp.id AS flex_policy_id,

        tp.earliest_start AS pol_earliest_start,
        tp.latest_start   AS pol_latest_start,
        tp.latest_end     AS pol_latest_end,
        tp.flex           AS pol_flex,

        0 AS grace_minutes
    FROM calendar c

    -- work_calendar_id nadal per user/dzień
    LEFT JOIN LATERAL (
        SELECT work_calendar_id
        FROM employee_assignments
        WHERE employee_id = @p_user_id
          AND c.day >= effective_from
          AND (effective_to IS NULL OR c.day <= effective_to)
        ORDER BY effective_from DESC, id DESC
        LIMIT 1
    ) ea ON TRUE

    -- time policy: wybór automatyczny po dniu (GLOBALNIE)
    LEFT JOIN LATERAL (
        SELECT id, earliest_start, latest_start, latest_end, flex
        FROM time_policies
        WHERE c.day >= effective_from
          AND (effective_to IS NULL OR c.day <= effective_to)
        ORDER BY effective_from DESC, id DESC
        LIMIT 1
    ) tp ON TRUE
),

-- ============================================================
-- 2) PLAN DNIA: policy < weekly < override
-- ============================================================
daily_plan AS (
    SELECT
        x.day,
        x.user_id,
        x.work_calendar_id,
        x.grace_minutes,

        o.id AS override_id,
        w.id AS weekly_id,

        COALESCE(o.planned_minutes, w.planned_minutes, 0) AS plan_minutes,

        COALESCE(o.start_time, o.earliest_start, w.earliest_start, x.pol_earliest_start) AS plan_start,
        COALESCE(o.end_time,   o.latest_end,    w.latest_end,    x.pol_latest_end)      AS plan_end,

        COALESCE(
          o.latest_start,
          w.latest_start,
          x.pol_latest_start,
          COALESCE(o.start_time, o.earliest_start, w.earliest_start, x.pol_earliest_start)
        ) AS latest_start,

        COALESCE(
          o.spans_midnight,
          w.spans_midnight,
          CASE
            WHEN COALESCE(o.start_time, o.earliest_start, w.earliest_start, x.pol_earliest_start) IS NOT NULL
             AND COALESCE(o.end_time,   o.latest_end,    w.latest_end,    x.pol_latest_end) IS NOT NULL
             AND COALESCE(o.start_time, o.earliest_start, w.earliest_start, x.pol_earliest_start)
               > COALESCE(o.end_time,   o.latest_end,    w.latest_end,    x.pol_latest_end)
            THEN 1 ELSE 0
          END
        ) AS spans_midnight,

        COALESCE(o.flex, w.flex, x.pol_flex, 1) AS flex,

        CASE
          WHEN o.id IS NOT NULL THEN 'override'
          WHEN w.id IS NOT NULL THEN 'weekly'
          WHEN x.pol_earliest_start IS NOT NULL OR x.pol_latest_end IS NOT NULL THEN 'policy'
          ELSE 'none'
        END AS plan_source
    FROM ctx x

    LEFT JOIN user_schedule_overrides o
      ON o.user_id = x.user_id AND o.day = x.day

    LEFT JOIN LATERAL (
        SELECT *
        FROM user_schedules_weekly
        WHERE user_id = x.user_id
          AND dow = (WEEKDAY(x.day) + 1)   -- Mon=1..Sun=7
          AND x.day >= effective_from
          AND (effective_to IS NULL OR x.day <= effective_to)
        ORDER BY effective_from DESC, id DESC
        LIMIT 1
    ) w ON TRUE
),

-- ============================================================
-- 3) ŚWIĘTA
-- ============================================================
daily_holidays AS (
    SELECT
        dp.day,
        dp.user_id,
        hol.id AS holiday_id,
        hol.name AS holiday_name
    FROM daily_plan dp
    LEFT JOIN calendar_holidays hol
      ON hol.day = dp.day AND hol.work_calendar_id = dp.work_calendar_id
),

-- ============================================================
-- 4) CHOROBOWE (1 rekord/dzień)
-- ============================================================
daily_sick AS (
    SELECT
        dp.day,
        dp.user_id,
        sick.id AS sick_id,
        sick.legal_basis,
        COALESCE(sick.is_occupational_accident, 0) AS sick_work_accident,
        COALESCE(sick.percent_coverage, 0) AS sick_percent_coverage
    FROM daily_plan dp
    LEFT JOIN LATERAL (
        SELECT id, legal_basis, is_occupational_accident, percent_coverage
        FROM employee_sickness
        WHERE employee_id = dp.user_id
          AND dp.day BETWEEN from_date AND to_date
        ORDER BY from_date DESC, id DESC
        LIMIT 1
    ) sick ON TRUE
),

-- ============================================================
-- 5) URLOPY dziennie (effective_minutes) – tylko approved
-- ============================================================
daily_leaves AS (
    SELECT
        lr.employee_id AS user_id,
        DATE(lri.start_at) AS leave_day,
        SUM(lri.effective_minutes) AS leave_minutes,

        MIN(lri.start_at) AS leave_first_start_at,
        MAX(lri.end_at)   AS leave_last_end_at,

        GROUP_CONCAT(DISTINCT lr.type_code ORDER BY lr.type_code SEPARATOR ', ') AS leave_type_codes,
        GROUP_CONCAT(DISTINCT lt.name ORDER BY lt.name SEPARATOR ', ') AS leave_names
    FROM leave_request_items lri
    JOIN leave_requests lr ON lr.id = lri.request_id
    JOIN leave_types lt ON lt.code = lr.type_code
    JOIN approval_requests ar
      ON ar.target_id = lr.id
     AND ar.resource = 'leave_requests'
     AND ar.status = 'approved'
    WHERE lr.employee_id = @p_user_id
      AND DATE(lri.start_at) BETWEEN @p_from_date AND @p_to_date
    GROUP BY lr.employee_id, DATE(lri.start_at)
),

-- ============================================================
-- 6) ROZLICZENIA – TABELI NIE MA -> ZERA, ale te same kolumny
-- ============================================================
daily_settlements AS (
    SELECT
        @p_user_id AS user_id,
        c.day AS work_day,
        0 AS bank_usage_minutes,
        0 AS payout_minutes,
        0 AS to_leave_minutes
    FROM calendar c
),

-- ============================================================
-- 7) PRACA RZECZYWISTA (logic_day = date(started_at - 4h))
-- ============================================================
work_metrics AS (
    SELECT
        user_id,
        DATE(started_at - INTERVAL 4 HOUR) AS logic_day,

        MIN(CASE WHEN type IN ('work','overtime') THEN started_at END) AS real_start_ts,
        MAX(CASE WHEN type IN ('work','overtime') THEN ended_at   END) AS real_end_ts,

        SUM(CASE WHEN type IN ('work','overtime')
                 THEN TIMESTAMPDIFF(MINUTE, started_at, IFNULL(ended_at, NOW()))
                 ELSE 0 END) AS work_minutes,

        SUM(CASE WHEN type='break'
                 THEN TIMESTAMPDIFF(MINUTE, started_at, IFNULL(ended_at, NOW()))
                 ELSE 0 END) AS break_minutes,

        SUM(
            CASE WHEN type IN ('work','overtime') THEN
              GREATEST(0, TIMESTAMPDIFF(MINUTE,
                GREATEST(started_at, CAST(CONCAT(DATE(started_at - INTERVAL 4 HOUR), ' 22:00:00') AS DATETIME)),
                LEAST(IFNULL(ended_at, NOW()), CAST(CONCAT(DATE(started_at - INTERVAL 4 HOUR) + INTERVAL 1 DAY, ' 06:00:00') AS DATETIME))
              ))
            ELSE 0 END
        ) AS night_minutes
    FROM timesheet_entries
    WHERE user_id = @p_user_id
      AND DATE(started_at - INTERVAL 4 HOUR) BETWEEN @p_from_date AND @p_to_date
    GROUP BY user_id, DATE(started_at - INTERVAL 4 HOUR)
),

-- ============================================================
-- 8) ODPOCZYNEK 11H
-- ============================================================
rest_analysis AS (
    SELECT
        wm.*,
        LAG(real_end_ts) OVER (PARTITION BY user_id ORDER BY logic_day) AS prev_end_ts,
        CASE
          WHEN LAG(real_end_ts) OVER (PARTITION BY user_id ORDER BY logic_day) IS NULL
            OR real_start_ts IS NULL
          THEN NULL
          ELSE TIMESTAMPDIFF(
                 MINUTE,
                 LAG(real_end_ts) OVER (PARTITION BY user_id ORDER BY logic_day),
                 real_start_ts
               )
        END AS rest_minutes
    FROM work_metrics wm
),

-- ============================================================
-- 9) IN/OUT JSON – z timesheet_entries (work/overtime)
-- ============================================================
spans AS (
    SELECT
        te.user_id,
        te.started_at,
        te.ended_at,
        DATE(te.started_at - INTERVAL 4 HOUR) AS logic_day,
        LEAD(te.started_at) OVER (PARTITION BY te.user_id ORDER BY te.started_at) AS next_started_at
    FROM timesheet_entries te
    WHERE te.user_id = @p_user_id
      AND te.type IN ('work','overtime')
      AND DATE(te.started_at - INTERVAL 4 HOUR) BETWEEN @p_from_date AND @p_to_date
),
spans_fixed AS (
    SELECT
        user_id,
        logic_day,
        started_at AS in_at,
        CASE
          WHEN ended_at IS NOT NULL THEN ended_at
          WHEN next_started_at IS NOT NULL THEN next_started_at
          ELSE NULL
        END AS out_at
    FROM spans
),
inout_by_day AS (
    SELECT
        user_id,
        logic_day,
        CAST(
            CONCAT(
                '[',
                IFNULL(
                    GROUP_CONCAT(
                        JSON_OBJECT(
                            'in',  DATE_FORMAT(in_at,  '%Y-%m-%d %H:%i:%s'),
                            'out', CASE WHEN out_at IS NULL THEN NULL
                                        ELSE DATE_FORMAT(out_at, '%Y-%m-%d %H:%i:%s')
                                   END
                        )
                        ORDER BY in_at SEPARATOR ','
                    ),
                    ''
                ),
                ']'
            ) AS JSON
        ) AS in_out
    FROM spans_fixed
    GROUP BY user_id, logic_day
),

-- ============================================================
-- 10) WYLICZENIA DNIA
-- ============================================================
calc AS (
    SELECT
        dp.day,
        dp.user_id,

        (WEEKDAY(dp.day) + 1) AS dow_mon1,
        DATE_FORMAT(dp.day, '%a') AS dow_name,

        dp.plan_source,
        CASE WHEN dp.flex=1 THEN 'ELASTYCZNY' ELSE 'SZTYWNY' END AS plan_type,

        dp.plan_minutes AS plan_minutes_raw,
        CASE
          WHEN dh.holiday_id IS NOT NULL THEN 0
          ELSE dp.plan_minutes
        END AS plan_minutes,

        dp.plan_start,
        dp.latest_start,
        dp.plan_end,
        dp.spans_midnight,
        dp.grace_minutes,

        CONCAT(
          COALESCE(TIME_FORMAT(dp.plan_start, '%H:%i'), ''),
          CASE WHEN dp.plan_start IS NOT NULL OR dp.plan_end IS NOT NULL THEN '-' ELSE '' END,
          COALESCE(TIME_FORMAT(dp.plan_end, '%H:%i'), '')
        ) AS schedule_hours,

        dh.holiday_id,
        dh.holiday_name,

        ds.sick_id,
        ds.legal_basis,
        COALESCE(ds.sick_work_accident, 0) AS sick_work_accident,
        COALESCE(ds.sick_percent_coverage, 0) AS sick_percent_coverage,

        COALESCE(dl.leave_minutes, 0) AS leave_minutes_raw,
        dl.leave_type_codes,
        dl.leave_names,

        dl.leave_first_start_at,
        dl.leave_last_end_at,

        COALESCE(dset.bank_usage_minutes, 0) AS bank_usage_minutes,
        COALESCE(dset.payout_minutes, 0)     AS payout_minutes,
        COALESCE(dset.to_leave_minutes, 0)   AS to_leave_minutes,

        COALESCE(ra.work_minutes, 0)   AS work_minutes,
        COALESCE(ra.break_minutes, 0)  AS break_minutes,
        COALESCE(ra.night_minutes, 0)  AS night_minutes,
        ra.real_start_ts,
        ra.real_end_ts,
        TIME_FORMAT(ra.real_start_ts, '%H:%i') AS real_start,
        TIME_FORMAT(ra.real_end_ts,   '%H:%i') AS real_end,
        ra.prev_end_ts,
        ra.rest_minutes,

        CASE
          WHEN (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END) <= 0 THEN 0
          WHEN COALESCE(dl.leave_minutes,0) >= (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END)
            THEN (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END)
          ELSE LEAST(COALESCE(dl.leave_minutes,0), (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END))
        END AS leave_credit_minutes,

        CASE
          WHEN (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END) <= 0 THEN 0
          WHEN ds.sick_id IS NOT NULL THEN (CASE WHEN dh.holiday_id IS NOT NULL THEN 0 ELSE dp.plan_minutes END)
          ELSE 0
        END AS sick_credit_minutes,

        0 AS holiday_credit_minutes
    FROM daily_plan dp
    LEFT JOIN daily_holidays dh ON dh.user_id=dp.user_id AND dh.day=dp.day
    LEFT JOIN daily_sick ds     ON ds.user_id=dp.user_id AND ds.day=dp.day
    LEFT JOIN daily_leaves dl   ON dl.user_id=dp.user_id AND dl.leave_day=dp.day
    LEFT JOIN daily_settlements dset ON dset.user_id=dp.user_id AND dset.work_day=dp.day
    LEFT JOIN rest_analysis ra  ON ra.user_id=dp.user_id AND ra.logic_day=dp.day
),

daily AS (
    SELECT
        c.*,

        LEAST(
          c.bank_usage_minutes,
          GREATEST(
            0,
            (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
            - c.work_minutes
          )
        ) AS bank_used_minutes,

        GREATEST(
          0,
          c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes)
          - LEAST(
              c.bank_usage_minutes,
              GREATEST(
                0,
                (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                - c.work_minutes
              )
            )
        ) AS required_minutes,

        LEAST(
          c.plan_minutes,
          c.work_minutes
          + c.leave_credit_minutes
          + c.sick_credit_minutes
          + c.holiday_credit_minutes
          + LEAST(
              c.bank_usage_minutes,
              GREATEST(
                0,
                (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                - c.work_minutes
              )
            )
        ) AS credited_to_norm_minutes,

        (c.work_minutes
          - GREATEST(
              0,
              c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes)
              - LEAST(
                  c.bank_usage_minutes,
                  GREATEST(
                    0,
                    (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                    - c.work_minutes
                  )
                )
            )
        ) AS gross_delta_minutes,

        CASE
          WHEN (c.work_minutes
                - GREATEST(
                    0,
                    c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes)
                    - LEAST(
                        c.bank_usage_minutes,
                        GREATEST(
                          0,
                          (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                          - c.work_minutes
                        )
                      )
                  )
               ) >= 0
          THEN GREATEST(
                 0,
                 (c.work_minutes
                   - GREATEST(
                       0,
                       c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes)
                       - LEAST(
                           c.bank_usage_minutes,
                           GREATEST(
                             0,
                             (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                             - c.work_minutes
                           )
                         )
                     )
                 ) - (c.payout_minutes + c.to_leave_minutes)
               )
          ELSE (c.work_minutes
                 - GREATEST(
                     0,
                     c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes)
                     - LEAST(
                         c.bank_usage_minutes,
                         GREATEST(
                           0,
                           (c.plan_minutes - (c.leave_credit_minutes + c.sick_credit_minutes + c.holiday_credit_minutes))
                           - c.work_minutes
                         )
                       )
                   )
               )
        END AS net_delta_minutes,

        CASE
          WHEN c.plan_type <> 'SZTYWNY' THEN 0
          WHEN c.plan_minutes <= 0 THEN 0
          WHEN c.real_start_ts IS NULL OR c.plan_start IS NULL THEN 0
          ELSE GREATEST(
                 0,
                 TIMESTAMPDIFF(
                   MINUTE,
                   CAST(CONCAT(c.day,' ',TIME_FORMAT(c.plan_start,'%H:%i:%s')) AS DATETIME),
                   c.real_start_ts
                 ) - c.grace_minutes
               )
        END AS late_minutes,

        CASE
          WHEN c.plan_type <> 'SZTYWNY' THEN 0
          WHEN c.plan_minutes <= 0 THEN 0
          WHEN c.real_end_ts IS NULL OR c.plan_end IS NULL THEN 0

          WHEN c.leave_credit_minutes > 0
               AND c.leave_last_end_at IS NOT NULL
               AND TIME(c.leave_last_end_at) = TIME(c.plan_end)
          THEN
            GREATEST(
              0,
              TIMESTAMPDIFF(
                MINUTE,
                c.real_end_ts,
                (
                  CASE
                    WHEN c.spans_midnight = 1
                    THEN CAST(CONCAT(c.day + INTERVAL 1 DAY,' ',TIME_FORMAT(c.plan_end,'%H:%i:%s')) AS DATETIME)
                    ELSE CAST(CONCAT(c.day,' ',TIME_FORMAT(c.plan_end,'%H:%i:%s')) AS DATETIME)
                  END
                ) - INTERVAL c.leave_credit_minutes MINUTE
              ) - c.grace_minutes
            )

          ELSE
            GREATEST(
              0,
              TIMESTAMPDIFF(
                MINUTE,
                c.real_end_ts,
                CASE
                  WHEN c.spans_midnight = 1
                  THEN CAST(CONCAT(c.day + INTERVAL 1 DAY,' ',TIME_FORMAT(c.plan_end,'%H:%i:%s')) AS DATETIME)
                  ELSE CAST(CONCAT(c.day,' ',TIME_FORMAT(c.plan_end,'%H:%i:%s')) AS DATETIME)
                END
              ) - c.grace_minutes
            )
        END AS early_leave_minutes,

        CASE
          WHEN c.rest_minutes IS NOT NULL AND c.rest_minutes < 660 THEN '!!! 11H !!!'
          ELSE 'OK'
        END AS rest_alert,

        CASE
          WHEN c.sick_id IS NOT NULL AND c.sick_work_accident = 1 THEN CONCAT('CHOROBOWE - WYPADEK (', c.legal_basis, ')')
          WHEN c.sick_id IS NOT NULL THEN CONCAT('CHOROBOWE (', c.legal_basis, ')')
          WHEN c.leave_names IS NOT NULL AND c.plan_minutes > 0 AND c.leave_credit_minutes = c.plan_minutes THEN CONCAT('URLOP: ', UPPER(c.leave_names))
          WHEN c.leave_names IS NOT NULL AND c.leave_credit_minutes > 0 THEN CONCAT('URLOP CZĘŚCIOWY: ', UPPER(c.leave_names))
          WHEN c.holiday_id IS NOT NULL THEN CONCAT('ŚWIĘTO: ', c.holiday_name)
          WHEN c.work_minutes > 0 THEN 'OBECNY'
          WHEN c.plan_minutes = 0 THEN 'WOLNE'
          WHEN c.day < CURDATE() THEN 'NIEOBECNY (NN)'
          ELSE 'PLAN'
        END AS status
    FROM calc c
),

-- ============================================================
-- 10b) SUMA PLANU W ZAKRESIE
-- ============================================================
scope_plan AS (
    SELECT
        user_id,
        SUM(plan_minutes) AS scope_plan_minutes
    FROM daily
    GROUP BY user_id
),

-- ============================================================
-- 11) TARGET = Nominals * FTE rozłożone po dniach planu
-- ============================================================
targets_raw AS (
    SELECT
        d.day,
        d.user_id,
        ed.fte_percent,
        nm.nominal_minutes_100,
        sp.scope_plan_minutes,

        CASE WHEN nm.nominal_minutes_100 IS NULL THEN 1 ELSE 0 END AS nominal_missing,

        MAX(CASE WHEN d.plan_minutes > 0 THEN d.day END)
          OVER (PARTITION BY d.user_id) AS last_planned_day,

        CAST(
          CASE
            WHEN nm.nominal_minutes_100 IS NULL THEN 0
            ELSE ROUND(
                   CAST(nm.nominal_minutes_100 AS DECIMAL(18,6)) * CAST(ed.fte_percent AS DECIMAL(18,6)),
                   0
                 )
          END AS SIGNED
        ) AS total_target_minutes,

        CASE
          WHEN nm.nominal_minutes_100 IS NULL THEN CAST(0 AS DECIMAL(18,6))
          WHEN sp.scope_plan_minutes <= 0 THEN CAST(0 AS DECIMAL(18,6))
          WHEN d.plan_minutes <= 0 THEN CAST(0 AS DECIMAL(18,6))
          ELSE
            (CAST(
               ROUND(CAST(nm.nominal_minutes_100 AS DECIMAL(18,6)) * CAST(ed.fte_percent AS DECIMAL(18,6)), 0)
             AS DECIMAL(18,6)) * CAST(d.plan_minutes AS DECIMAL(18,6)))
            / CAST(sp.scope_plan_minutes AS DECIMAL(18,6))
        END AS target_share_raw
    FROM daily d
    JOIN et_day ed ON ed.user_id = d.user_id AND ed.day = d.day
    JOIN scope_plan sp ON sp.user_id = d.user_id
    CROSS JOIN month_meta mm
    LEFT JOIN nom_month nm ON TRUE
),

targets_int AS (
    SELECT
        tr.day,
        tr.user_id,
        tr.fte_percent,
        tr.nominal_minutes_100,
        tr.scope_plan_minutes,
        tr.nominal_missing,
        tr.last_planned_day,
        tr.total_target_minutes,

        CAST(FLOOR(tr.target_share_raw) AS SIGNED) AS target_base_minutes,

        CAST(
          (tr.total_target_minutes
            - SUM(CAST(FLOOR(tr.target_share_raw) AS SIGNED)) OVER (PARTITION BY tr.user_id)
          ) AS SIGNED
        ) AS remainder_minutes
    FROM targets_raw tr
),

targets AS (
    SELECT
        ti.day,
        ti.user_id,
        ti.fte_percent,
        ti.nominal_minutes_100,
        ti.scope_plan_minutes,
        ti.nominal_missing,

        CASE
          WHEN ti.nominal_missing = 1 THEN 0
          WHEN ti.last_planned_day IS NULL THEN 0
          WHEN ti.target_base_minutes = 0 AND ti.day <> ti.last_planned_day THEN 0
          WHEN ti.day = ti.last_planned_day THEN (ti.target_base_minutes + ti.remainder_minutes)
          ELSE ti.target_base_minutes
        END AS target_share_minutes
    FROM targets_int ti
),

daily_with_cum AS (
    SELECT
        d.*,
        io.in_out,

        t.fte_percent,
        t.nominal_minutes_100,
        mm.days_in_scope,
        t.nominal_missing,
        t.target_share_minutes,

        SUM(d.plan_minutes) OVER (ORDER BY d.day) AS cum_plan_minutes,
        SUM(d.credited_to_norm_minutes) OVER (ORDER BY d.day) AS cum_exec_minutes,
        SUM(d.net_delta_minutes) OVER (ORDER BY d.day) AS cum_net_delta_minutes,
        SUM(t.target_share_minutes) OVER (ORDER BY d.day) AS cum_target_minutes,

        (SUM(d.credited_to_norm_minutes) OVER (ORDER BY d.day)
         - SUM(d.plan_minutes) OVER (ORDER BY d.day)
        ) AS cum_exec_minus_plan_minutes,

        (SUM(d.credited_to_norm_minutes) OVER (ORDER BY d.day)
         - SUM(t.target_share_minutes) OVER (ORDER BY d.day)
        ) AS cum_exec_minus_target_minutes
    FROM daily d
    CROSS JOIN month_meta mm
    LEFT JOIN inout_by_day io ON io.user_id=d.user_id AND io.logic_day=d.day
    JOIN targets t ON t.user_id=d.user_id AND t.day=d.day
),

month_totals AS (
    SELECT
        user_id,
        SUM(plan_minutes) AS month_plan_minutes,
        SUM(credited_to_norm_minutes) AS month_exec_minutes,
        SUM(net_delta_minutes) AS month_net_delta_minutes,
        SUM(target_share_minutes) AS month_target_minutes,
        MAX(nominal_missing) AS nominal_missing_any
    FROM daily_with_cum
    GROUP BY user_id
)

SELECT
    d.day,
    d.user_id,
    xx.name,
    xx.surname,
    d.dow_mon1 AS dow,
    d.dow_name,

    d.plan_source,
    d.plan_type,
    d.schedule_hours,

    d.plan_minutes_raw,
    d.plan_minutes,

    d.required_minutes,

    d.leave_type_codes,
    d.leave_names,
    d.leave_minutes_raw,
    d.leave_credit_minutes,

    DATE_FORMAT(d.leave_first_start_at, '%Y-%m-%d %H:%i:%s') AS leave_first_start_at,
    DATE_FORMAT(d.leave_last_end_at,  '%Y-%m-%d %H:%i:%s') AS leave_last_end_at,

    d.sick_id,
    d.legal_basis,
    d.sick_work_accident,
    d.sick_percent_coverage,

    d.holiday_name,

    d.work_minutes,
    d.break_minutes,
    d.night_minutes,
    d.real_start,
    d.real_end,

    d.bank_usage_minutes,
    d.bank_used_minutes,
    d.payout_minutes,
    d.to_leave_minutes,

    d.credited_to_norm_minutes,
    d.gross_delta_minutes,
    d.net_delta_minutes,

    d.late_minutes,
    d.early_leave_minutes,

    d.rest_minutes,
    d.rest_alert,
    d.status,

    d.fte_percent,
    d.nominal_minutes_100,
    d.nominal_missing,
    d.target_share_minutes,

    d.cum_target_minutes,
    d.cum_plan_minutes,
    d.cum_exec_minutes,
    d.cum_net_delta_minutes,
    d.cum_exec_minus_plan_minutes,
    d.cum_exec_minus_target_minutes,

    mt.month_plan_minutes,
    mt.month_exec_minutes,
    mt.month_net_delta_minutes,
    mt.month_target_minutes,
    (mt.month_exec_minutes - mt.month_plan_minutes) AS month_exec_minus_plan_minutes,
    (mt.month_exec_minutes - mt.month_target_minutes) AS month_exec_minus_target_minutes,
    mt.nominal_missing_any,

    d.in_out
FROM daily_with_cum d
JOIN month_totals mt ON mt.user_id = d.user_id
LEFT JOIN users xx ON d.user_id = xx.id
ORDER BY d.day;


`

	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return du.MonthlyReport{}, err
	}
	defer rows.Close()

	report := du.MonthlyReport{
		UserID:   userID,
		FromDate: fromDate.Format("2006-01-02"),
		ToDate:   toDate.Format("2006-01-02"),
		Days:    make([]du.DayReportRow, 0, 32),
	}

	var gotAny bool

	for rows.Next() {
		gotAny = true

		// Nullables
		var (
			leaveTypeCodes sql.NullString
			leaveNames     sql.NullString
leaveFirst sql.NullString
leaveLast  sql.NullString
			sickID         sql.NullInt64
			legalBasis     sql.NullString
			holidayName    sql.NullString
			realStart      sql.NullString
			realEnd        sql.NullString
			restMinutes    sql.NullInt64
			inOut          sql.NullString

			// Totals (powtarzają się w każdym wierszu)
			mtPlan, mtExec, mtNet, mtTarget                 int
			mtExecMinusPlan, mtExecMinusTarget, mtNomMissAny int
		)

		var day du.DayReportRow

		if err := rows.Scan(
			&day.Day,
			&day.UserID,
			&day.Name,
			&day.Surname,
			&day.Dow,
			&day.DowName,
			&day.PlanSource,
			&day.PlanType,
			&day.ScheduleHours,
			&day.PlanMinutesRaw,
			&day.PlanMinutes,
			&day.RequiredMinutes,
			&leaveTypeCodes,
			&leaveNames,
			&day.LeaveMinutesRaw,
			&day.LeaveCreditMinutes,
			&leaveFirst,
			&leaveLast,
			&sickID,
			&legalBasis,
			&day.SickWorkAccident,
			&day.SickPercentCoverage,
			&holidayName,
			&day.WorkMinutes,
			&day.BreakMinutes,
			&day.NightMinutes,
			&realStart,
			&realEnd,
			&day.BankUsageMinutes,
			&day.BankUsedMinutes,
			&day.PayoutMinutes,
			&day.ToLeaveMinutes,
			&day.CreditedToNormMinutes,
			&day.GrossDeltaMinutes,
			&day.NetDeltaMinutes,
			&day.LateMinutes,
			&day.EarlyLeaveMinutes,
			&restMinutes,
			&day.RestAlert,
			&day.Status,
			&day.FtePercent,
			&day.NominalMinutes100,
			&day.NominalMissing,
			&day.TargetShareMinutes,
			&day.CumTargetMinutes,
			&day.CumPlanMinutes,
			&day.CumExecMinutes,
			&day.CumNetDeltaMinutes,
			&day.CumExecMinusPlanMinutes,
			&day.CumExecMinusTargetMinutes,
			&mtPlan,
			&mtExec,
			&mtNet,
			&mtTarget,
			&mtExecMinusPlan,
			&mtExecMinusTarget,
			&mtNomMissAny,
			&inOut,
		); err != nil {
			return du.MonthlyReport{}, err
		}

		// Map nullables -> pointers
		if leaveTypeCodes.Valid {
			v := leaveTypeCodes.String
			day.LeaveTypeCodes = &v
		}
		if leaveNames.Valid {
			v := leaveNames.String
			day.LeaveNames = &v
		}
if leaveFirst.Valid && leaveFirst.String != "" {
    v := leaveFirst.String
    day.LeaveFirstStartAt = &v
}
if leaveLast.Valid && leaveLast.String != "" {
    v := leaveLast.String
    day.LeaveLastEndAt = &v
}

		if sickID.Valid {
			v := int(sickID.Int64)
			day.SickID = &v
		}
		if legalBasis.Valid {
			v := legalBasis.String
			day.LegalBasis = &v
		}
		if holidayName.Valid {
			v := holidayName.String
			day.HolidayName = &v
		}
		if realStart.Valid {
			v := realStart.String
			day.RealStart = &v
		}
		if realEnd.Valid {
			v := realEnd.String
			day.RealEnd = &v
		}
		if restMinutes.Valid {
			v := int(restMinutes.Int64)
			day.RestMinutes = &v
		}
		if inOut.Valid && inOut.String != "" {
			day.InOut = json.RawMessage(inOut.String)
		}

		// Totals (wystarczy nadpisywać – i tak stałe w obrębie raportu)
		report.Totals = du.MonthlyTotals{
			MonthPlanMinutes:            mtPlan,
			MonthExecMinutes:            mtExec,
			MonthNetDeltaMinutes:        mtNet,
			MonthTargetMinutes:          mtTarget,
			MonthExecMinusPlanMinutes:   mtExecMinusPlan,
			MonthExecMinusTargetMinutes: mtExecMinusTarget,
			NominalMissingAny:           mtNomMissAny,
		}

		// Dane użytkownika (też stałe)
		report.Name = day.Name
		report.Surname = day.Surname

		report.Days = append(report.Days, day)
	}

	if err := rows.Err(); err != nil {
		return du.MonthlyReport{}, err
	}

	// CTE calendar gwarantuje wiersze, ale zabezpieczenie zostawiamy.
	if !gotAny {
		report.Days = []du.DayReportRow{}
	}

	return report, nil
}
