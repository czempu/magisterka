package mysql

import (
	"context"
	"database/sql"
	"strings"
	du "erpbadger/internal/domain/workflow"
)

type WorkflowRepo struct {
	db *sql.DB
}

func NewWorkflowRepo(db *sql.DB) du.Repository {
	return &WorkflowRepo{db: db}
}
func (r *WorkflowRepo) DeleteStepsNotIn(ctx context.Context, wfID int, stepNos []int) error {
	if len(stepNos) == 0 {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM approval_workflow_steps WHERE workflow_id = ?`,
			wfID,
		)
		return err
	}

	query := `
		DELETE FROM approval_workflow_steps
		WHERE workflow_id = ?
		AND step_no NOT IN (?` + strings.Repeat(",?", len(stepNos)-1) + `)
	`

	args := []any{wfID}
	for _, n := range stepNos {
		args = append(args, n)
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *WorkflowRepo) GetWorkflows(ctx context.Context) ([]du.Workflow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, resource FROM approval_workflows`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []du.Workflow
	for rows.Next() {
		var wf du.Workflow
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.Resource); err != nil {
			return nil, err
		}
		res = append(res, wf)
	}
	return res, nil
}
func (r *WorkflowRepo) UpsertStep(ctx context.Context, s du.WorkflowStep) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO approval_workflow_steps
		(workflow_id, step_no, completion_rule, approver_type, slug, levels_up, org_scope, required)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			completion_rule = VALUES(completion_rule),
			approver_type   = VALUES(approver_type),
			slug            = VALUES(slug),
			levels_up       = VALUES(levels_up),
			org_scope       = VALUES(org_scope),
			required        = VALUES(required)
	`,
		s.WorkflowID,
		s.StepNo,
		s.CompletionRule,
		s.ApproverType,
		s.Slug,
		s.LevelsUp,
		s.OrgScope,
		s.Required,
	)
	return err
}

func (r *WorkflowRepo) GetWorkflowForResource(ctx context.Context, resource string) (du.Workflow, error) {
	var wf du.Workflow

	err := r.db.QueryRowContext(ctx, `
		SELECT w.id, w.name, w.resource
		FROM approval_workflows w
		JOIN approval_policies p ON p.workflow_id = w.id
		WHERE p.resource = ?
		LIMIT 1`, resource,
	).Scan(&wf.ID, &wf.Name, &wf.Resource)
	if err != nil {
		return wf, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT workflow_id, step_no, completion_rule, approver_type, slug, levels_up, org_scope, required
		FROM approval_workflow_steps
		WHERE workflow_id=?
		ORDER BY step_no`, wf.ID)
	if err != nil {
		return wf, err
	}
	defer rows.Close()

	for rows.Next() {
		var s du.WorkflowStep
		if err := rows.Scan(
			&s.WorkflowID,
			&s.StepNo,
			&s.CompletionRule,
			&s.ApproverType,
			&s.Slug,
			&s.LevelsUp,
			&s.OrgScope,
			&s.Required,
		); err != nil {
			return wf, err
		}
		wf.Steps = append(wf.Steps, s)
	}

	return wf, nil
}

func (r *WorkflowRepo) CreateWorkflow(ctx context.Context, name, resource string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO approval_workflows (name, resource) VALUES (?, ?)`,
		name, resource)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *WorkflowRepo) CreatePolicy(ctx context.Context, resource string, workflowID int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO approval_policies (resource, workflow_id) VALUES (?, ?)`,
		resource, workflowID)
	return err
}

func (r *WorkflowRepo) AddStep(ctx context.Context, s du.WorkflowStep) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO approval_workflow_steps
		(workflow_id, step_no, completion_rule, approver_type, slug, levels_up, org_scope, required)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.WorkflowID,
		s.StepNo,
		s.CompletionRule,
		s.ApproverType,
		s.Slug,
		s.LevelsUp,
		s.OrgScope,
		s.Required,
	)
	return err
}
func (r *WorkflowRepo) GetWorkflowByID(ctx context.Context, id int) (du.Workflow, error) {
	var wf du.Workflow

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, resource
		FROM approval_workflows
		WHERE id = ?`, id,
	).Scan(&wf.ID, &wf.Name, &wf.Resource)
	if err != nil {
		return wf, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT workflow_id, step_no, completion_rule, approver_type, slug, levels_up, org_scope, required
		FROM approval_workflow_steps
		WHERE workflow_id = ?
		ORDER BY step_no`, id)
	if err != nil {
		return wf, err
	}
	defer rows.Close()

	for rows.Next() {
		var s du.WorkflowStep
		if err := rows.Scan(
			&s.WorkflowID,
			&s.StepNo,
			&s.CompletionRule,
			&s.ApproverType,
			&s.Slug,
			&s.LevelsUp,
			&s.OrgScope,
			&s.Required,
		); err != nil {
			return wf, err
		}
		wf.Steps = append(wf.Steps, s)
	}

	return wf, nil
}

