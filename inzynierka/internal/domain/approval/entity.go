package approval

import "time"

type ApprovalRequest struct {
	ID         int64     `json:"id"`
	Resource   string    `json:"resource"`
	TargetID   int64     `json:"target_id"`
	Requester  int       `json:"requester_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	ClosedAt   *time.Time `json:"closed_at"`
}

type ApprovalStep struct {
	ID         int64      `json:"id"`
	RequestID  int64      `json:"request_id"`
	StepNo     int        `json:"step_no"`
	ApproverID *int        `json:"approver_id"`
	Status     string     `json:"status"`
	DecidedAt  *time.Time `json:"decided_at"`
	Note       *string    `json:"note"`
	Approver_type string `json:"approver_type"`
	Approver_org_unitID *int `json:"approver_org_unit_id"`
	Approver_role_code *int `json:"approver_role_code"`
}
type ApprovalStepList struct {
	ID         int64      `json:"id"`
	RequestID  int64      `json:"request_id"`
	StepNo     int        `json:"step_no"`
	ApproverID *int        `json:"approver_id"`
	Status     string     `json:"status"`
	TargetID   int
	Resource   string 
	Name	   *string
	Surname	   *string
	LeaveRequestID int
	EmployeeID int
	TypeCode string
	CreatedAt  *string 
	LeaveStart *string
	LeaveEnd *string
	Type *string
}

