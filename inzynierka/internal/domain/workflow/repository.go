package workflow

import "context"

type Repository interface {
	GetWorkflows(ctx context.Context) ([]Workflow, error)
	GetWorkflowByID(ctx context.Context, id int) (Workflow, error)
	GetWorkflowForResource(ctx context.Context, resource string) (Workflow, error)
	CreateWorkflow(ctx context.Context, name, resource string) (int64, error)
	CreatePolicy(ctx context.Context, resource string, workflowID int) error
	UpsertStep(ctx context.Context, step WorkflowStep) error
	DeleteStepsNotIn(ctx context.Context, workflowID int, stepNos []int) error
}

