package mysql

import (
	"context"
	"database/sql"
	"time"
	"fmt"
	du "erpbadger/internal/domain/employment"
)

type EmploymentRepo struct {
	db *sql.DB
}

func NewEmploymentRepo(db *sql.DB) du.Repository {
	return &EmploymentRepo{db: db}
}

func (r *EmploymentRepo) GetEmploymentTerms(ctx context.Context, userID int) ([]du.EmploymentTerm, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, effective_from, effective_to, fte_percent, position
		FROM employment_terms WHERE user_id = ? ORDER BY effective_from DESC;`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []du.EmploymentTerm
	for rows.Next() {
		var t du.EmploymentTerm
		if err := rows.Scan(&t.ID, &t.UserID, &t.EffectiveFrom, &t.EffectiveTo, &t.FTEPercent, &t.Position); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *EmploymentRepo) AddEmploymentTerm(ctx context.Context, term du.EmploymentTermAdd) error {

if term.EffectiveFrom == "" {
		return fmt.Errorf("musi być określony początek obowiązywania grafiku (effective_from)")
	}

	// popraw format daty, nawet jeśli przyszło z timestampem
	var fromDate string
	{
		val := term.EffectiveFrom
		if len(val) > 10 {
			val = val[:10]
		}
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return fmt.Errorf("niepoprawny format effective_from: %w", err)
		}
		fromDate = t.Format("2006-01-02")
	}

	// ---- effective_to: MOŻE być NULL ----
	var toDate sql.NullString
	if term.EffectiveTo != nil && *term.EffectiveTo != "" {
		val := *term.EffectiveTo
		if len(val) > 10 {
			val = val[:10]
		}
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return fmt.Errorf("niepoprawny format effective_to: %w", err)
		}
		toDate = sql.NullString{String: t.Format("2006-01-02"), Valid: true}
	} else {
		toDate = sql.NullString{Valid: false}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO employment_terms (user_id, effective_from, effective_to, fte_percent, position)
		VALUES (?, ?, ?, ?, ?);`,
		term.UserID, fromDate, toDate, term.FTEPercent, term.Position)
	return err
}

func (r *EmploymentRepo) UpdateEmploymentTerm(ctx context.Context, term du.EmploymentTerm) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE employment_terms 
		SET fte_percent=?, position=?, effective_from = ?, effective_to=?
		WHERE id=?;`,
		term.FTEPercent,  term.Position, term.EffectiveFrom, term.EffectiveTo, term.ID)
	return err
}


func (r *EmploymentRepo) DeleteEmploymentTerm(ctx context.Context, id int) error {
  query := "delete from employment_terms where id =?"
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {   
    	return err
    }
    return nil
}


func (r *EmploymentRepo) EmploymentTermExist(ctx context.Context, userid int, id *int, from string, to *string) (bool,error) {
fmt.Println(userid)

fmt.Println(from)
fmt.Println(to)

if id == nil {
	fmt.Println("1")
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM employment_terms e
  WHERE ? <= COALESCE(e.effective_to, '9999-12-31')
    AND COALESCE(?, '9999-12-31') >= e.effective_from
    and e.user_id = ?
) AS is_collision;
 `

	var exists bool
	if err := r.db.QueryRowContext(ctx, q,from, to, userid).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

	fmt.Println("2")
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM employment_terms e
  WHERE ? <= COALESCE(e.effective_to, '9999-12-31')
    AND COALESCE(?, '9999-12-31') >= e.effective_from
    and e.user_id = ?
    and id <> ?
) AS is_collision;
 `

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, from, to, userid, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil

}
