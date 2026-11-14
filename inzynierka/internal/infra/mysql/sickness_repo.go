package mysql

import (
	"context"
	"database/sql"
	"fmt"

	du "erpbadger/internal/domain/sickness"
)

type SicknessRepo struct {
	db *sql.DB
}

func NewSicknessRepo(db *sql.DB) du.Repository {
	return &SicknessRepo{db: db}
}


func (r *SicknessRepo) SickExist(ctx context.Context,  s du.Sickness) (bool,error) {


	const q = `
		SELECT EXISTS (
  SELECT 1
  FROM employee_sickness e
  WHERE ? <= COALESCE(e.to_date, '9999-12-31')
    AND ? >= e.from_date
    and e.employee_id = ?
) AS is_collision; `

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, s.FromDate, s.ToDate, s.EmployeeID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil

}









// LIST
func (r *SicknessRepo) List(
	ctx context.Context,
	employeeID *int,
	status *string,
) ([]du.Sickness, error) {
	q := `
		SELECT
			a.id,
			employee_id,
			from_date,
			to_date,
			legal_basis,
			certificate_number,
			percent_coverage,
			is_occupational_accident,
			ar.status,
			reported_at,
			a.created_at,
			updated_at
		FROM approval_requests ar
		left join employee_sickness a on ar.resource = 'sick' and ar.target_id = a.id 
		WHERE 1=1
	`

	args := []any{}

	if employeeID != nil && *employeeID > 0 {
		q += " AND employee_id = ?"
		args = append(args, *employeeID)
	}
	if status != nil && *status != "" {
		q += " AND status = ?"
		args = append(args, *status)
	}

	q += " ORDER BY from_date DESC, id DESC"
	fmt.Println(q)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []du.Sickness
	for rows.Next() {
		var s du.Sickness
		var cert sql.NullString

		if err := rows.Scan(
			&s.ID,
			&s.EmployeeID,
			&s.FromDate,
			&s.ToDate,
			&s.LegalBasis,
			&cert,
			&s.PercentCoverage,
			&s.IsOccupationalAccident,
			&s.Status,
			&s.ReportedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if cert.Valid {
			s.CertificateNumber = cert.String
		}

		res = append(res, s)
	}
	return res, rows.Err()
}

// GET BY ID
func (r *SicknessRepo) GetByID(ctx context.Context, id int) (*du.Sickness, error) {
	q := `
		SELECT
			a.id,
			employee_id,
			from_date,
			to_date,
			legal_basis,
			certificate_number,
			percent_coverage,
			is_occupational_accident,
			ar.status,
			reported_at,
			a.created_at,
			updated_at
		FROM employee_sickness a
		left join approval_requests ar on ar.resource = 'sick' and ar.target_id = a.id 
		WHERE id = ?
		LIMIT 1
	`

	var s du.Sickness
	var cert sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&s.ID,
		&s.EmployeeID,
		&s.FromDate,
		&s.ToDate,
		&s.LegalBasis,
		&cert,
		&s.PercentCoverage,
		&s.IsOccupationalAccident,
		&s.Status,
		&s.ReportedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if cert.Valid {
		s.CertificateNumber = cert.String
	}

	return &s, nil
}

// CREATE
func (r *SicknessRepo) Create(ctx context.Context, s du.Sickness) (int, error) {
	q := `
		INSERT INTO employee_sickness (
			employee_id,
			from_date,
			to_date,
			legal_basis,
			certificate_number,
			percent_coverage,
			is_occupational_accident,
			status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := r.db.ExecContext(ctx, q,
		s.EmployeeID,
		s.FromDate,
		s.ToDate,
		s.LegalBasis,
		sql.NullString{String: s.CertificateNumber, Valid: s.CertificateNumber != ""},
		s.PercentCoverage,
		s.IsOccupationalAccident,
		s.Status,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Create Sickness LastInsertId: %w", err)
	}
	return int(id), nil
}

// DELETE
func (r *SicknessRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM employee_sickness WHERE id = ?`, id)
	if err != nil {
	return err
	}
	_, err2 := r.db.ExecContext(ctx, `delete from approval_requests ar where ar.resource = 'sick' or ar.target_id = ?`, id)
	if err2 != nil {
	return err2
	}
	return err
}

func (r *SicknessRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE employee_sickness SET status = ? WHERE id = ?`,
		status, id,
	)
	return err
}

