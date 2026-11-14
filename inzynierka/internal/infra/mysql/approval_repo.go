package mysql

import (
	"context"
	"database/sql"
	"erpbadger/internal/domain/approval"
)

type ApprovalRepo struct {
	db *sql.DB
}

func NewApprovalRepo(db *sql.DB) approval.Repository {
	return &ApprovalRepo{db: db}
}

func (r *ApprovalRepo) CreateRequest(ctx context.Context, req approval.ApprovalRequest) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO approval_requests (resource, target_id, requester_id, status, created_at)
		VALUES (?, ?, ?, ?, NOW())`,
		req.Resource, req.TargetID, req.Requester, req.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ApprovalRepo) CreateStep(ctx context.Context, step approval.ApprovalStep) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO approval_request_steps (request_id, step_no, approver_id, status,approver_type,approver_org_unit_id, approver_role_id)
		VALUES (?, ?, ?, ?,?,?,?)`,
		step.RequestID, step.StepNo, step.ApproverID, step.Status, step.Approver_type, step.Approver_org_unitID ,step.Approver_role_code)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ApprovalRepo) UpdateStepStatus(ctx context.Context, stepID int64, status string, note *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE approval_request_steps SET status=?, decided_at=NOW(), note=? WHERE id=?`,
		status, note, stepID)
	return err
}



func (r *ApprovalRepo) UpdateStepStatusAndApproverID(ctx context.Context, stepID int64, status string, note *string, ApproverID int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE approval_request_steps SET status=?, decided_at=NOW(), note=?, approver_id = ?  WHERE id=?`,
		status, note,ApproverID, stepID)
	return err
}
func (r *ApprovalRepo) UpdateRequestStatus(ctx context.Context, requestID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE approval_requests SET status=?, closed_at=NOW() WHERE id=?`, status, requestID)
	return err
}

func (r *ApprovalRepo) GetPendingStepsForApprover(ctx context.Context, approverID int) ([]approval.ApprovalStepList, error) {
	rows, err := r.db.QueryContext(ctx, `
		with leaves as (
select lri.request_id, min(lri.start_at ) leave_start, max(end_at) "leave_end" from leave_request_items lri  group by lri.request_id 
)
select x.*, u.name, u.surname, lr.*, y.leave_start , y.leave_end, lt.name FROM 
	(SELECT a.id, request_id, step_no, approver_id, a.status , ar.target_id, ar.resource  FROM approval_request_steps a
			left join approval_requests ar on request_id = ar.id
		WHERE a.status='pending' and (approver_id=?  or approver_role_id = (select COALESCE(role_id,0) as role_id  from user_roles WHERE user_id  = ?) 
		or approver_org_unit_id = (SELECT COALESCE(org_unit_id,0) as org_unit_id from user_org_memberships 
		where user_id = ? and since_date < now() and (until_date < now() or until_date is null)))) x
		left join leave_requests lr on x.target_id = lr.id  
		left join users u on lr.employee_id = u.id
		left join leaves y on x.target_id = y.request_id
		left join leave_types lt on lt.code = lr.type_code 
		`, approverID, approverID,approverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []approval.ApprovalStepList
	for rows.Next() {
		var s approval.ApprovalStepList
		if err := rows.Scan(
		&s.ID,
		&s.RequestID, 
		&s.StepNo, 
		&s.ApproverID, 
		&s.Status,
		&s.TargetID,
		&s.Resource,
		&s.Name,
		&s.Surname,
		&s.LeaveRequestID,
		&s.EmployeeID,
		&s.TypeCode,
		&s.CreatedAt,
		&s.LeaveStart,
		&s.LeaveEnd,
		&s.Type); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

func (r *ApprovalRepo) GetRequestDetails(ctx context.Context, requestID int64) (approval.ApprovalRequest, error) {
	var a approval.ApprovalRequest
	err := r.db.QueryRowContext(ctx, `
		SELECT id, resource, target_id, requester_id, status, created_at, closed_at
		FROM approval_requests WHERE id=?`, requestID).Scan(
		&a.ID, &a.Resource, &a.TargetID, &a.Requester, &a.Status, &a.CreatedAt, &a.ClosedAt)
	return a, err
}


func (r *ApprovalRepo) GetStepByID(ctx context.Context, stepID int64) (approval.ApprovalStep, error) {
	var s approval.ApprovalStep
	err := r.db.QueryRowContext(ctx, `
		SELECT id, request_id, step_no, approver_id, status,note,approver_type,approver_org_unit_id, approver_role_id
		FROM approval_request_steps WHERE id=?`, stepID).
		Scan(&s.ID, &s.RequestID, &s.StepNo, &s.ApproverID, &s.Status, &s.Note, &s.Approver_type, &s.Approver_org_unitID, &s.Approver_role_code)
	return s, err
}

func (r *ApprovalRepo) GetNextStep(ctx context.Context, requestID int64, currentStepNo int) (approval.ApprovalStep, error) {
	var s approval.ApprovalStep
	err := r.db.QueryRowContext(ctx, `
		SELECT id, request_id, step_no, approver_id, status
		FROM approval_request_steps
		WHERE request_id=? AND step_no > ? ORDER BY step_no ASC LIMIT 1`, requestID, currentStepNo).
		Scan(&s.ID, &s.RequestID, &s.StepNo, &s.ApproverID, &s.Status)
	return s, err
}

func (r *ApprovalRepo) UpdateStepStatusByID(ctx context.Context, stepID int64, status string, note *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE approval_request_steps SET status=?, decided_at=NOW(), note=? WHERE id=?`, status, note, stepID)
	return err
}


