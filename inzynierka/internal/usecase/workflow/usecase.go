package workflow

import (
	"context"
	"errors"

	du "erpbadger/internal/domain/workflow"
)

type Usecase interface {
GetWorkflows(ctx context.Context) ([]du.Workflow, error)
	GetWorkflowByID(ctx context.Context, id int) (du.Workflow, error)
	GetWorkflowForResource(ctx context.Context, resource string) (du.Workflow, error)

	CreateWorkflow(ctx context.Context, name, resource string) error
	UpdateWorkflowSteps(ctx context.Context, workflowID int, steps []du.WorkflowStep) error
}

type usecase struct {
	repo du.Repository
}

func NewUsecase(r du.Repository) Usecase {
	return &usecase{repo: r}
}
func (u *usecase) GetWorkflows(ctx context.Context) ([]du.Workflow, error) {
	return u.repo.GetWorkflows(ctx)
}
func (u *usecase) GetWorkflowByID(ctx context.Context, id int) (du.Workflow, error) {
	if id <= 0 {
		return du.Workflow{}, errors.New("invalid workflow id")
	}
	return u.repo.GetWorkflowByID(ctx, id)
}

func (u *usecase) GetWorkflowForResource(ctx context.Context, resource string) (du.Workflow, error) {
	if resource == "" {
		return du.Workflow{}, errors.New("resource required")
	}
	return u.repo.GetWorkflowForResource(ctx, resource)
}

func (u *usecase) CreateWorkflow(ctx context.Context, name, resource string) error {
	if name == "" || resource == "" {
		return errors.New("name and resource required")
	}

	id, err := u.repo.CreateWorkflow(ctx, name, resource)
	if err != nil {
		return err
	}

	return u.repo.CreatePolicy(ctx, resource, int(id))
}
func (u *usecase) UpdateWorkflowSteps(
	ctx context.Context,
	workflowID int,
	steps []du.WorkflowStep,
) error {

	if workflowID <= 0 {
		return errors.New("invalid workflow id")
	}

	stepNos := make([]int, 0, len(steps))

	for _, s := range steps {
		if s.StepNo < 1 {
			return errors.New("step_no must be >= 1")
		}

		switch s.ApproverType {
		case "role_global", "manager_of_unit", "user_of_unit", "specific_user":
			if s.Slug == nil {
				return errors.New("role_slug required for approver_type")
			}
		default:
			s.Slug = nil
		}

		s.WorkflowID = workflowID
		stepNos = append(stepNos, s.StepNo)

		if err := u.repo.UpsertStep(ctx, s); err != nil {
			return err
		}
	}

	return u.repo.DeleteStepsNotIn(ctx, workflowID, stepNos)
}

