package approval

import "context"

type Repository interface {
	CreateRequest(ctx context.Context, req ApprovalRequest) (int64, error)
	CreateStep(ctx context.Context, step ApprovalStep) (int64, error)
	UpdateStepStatus(ctx context.Context, stepID int64, status string, note *string) error
	UpdateRequestStatus(ctx context.Context, requestID int64, status string) error
	GetPendingStepsForApprover(ctx context.Context, approverID int) ([]ApprovalStepList, error)
	GetRequestDetails(ctx context.Context, requestID int64) (ApprovalRequest, error)
UpdateStepStatusAndApproverID(ctx context.Context, stepID int64, status string, note *string, ApproverID int) error
	GetStepByID(ctx context.Context, stepID int64) (ApprovalStep, error)
	GetNextStep(ctx context.Context, requestID int64, currentStepNo int) (ApprovalStep, error)
	UpdateStepStatusByID(ctx context.Context, stepID int64, status string, note *string) error
}

