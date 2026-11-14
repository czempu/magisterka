package workflow

type Workflow struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Resource string          `json:"resource"`
	Steps    []WorkflowStep  `json:"steps"`
}

type WorkflowStep struct {
	WorkflowID     int    `json:"workflow_id"`
	StepNo         int    `json:"step_no"`
	CompletionRule string `json:"completion_rule"`
	ApproverType   string `json:"approver_type"`
	Slug       *string `json:"role_slug"`
	LevelsUp       int    `json:"levels_up"`
	OrgScope       string `json:"org_scope"`
	Required       bool   `json:"required"`
}

type Policy struct {
	ID         int    `json:"id"`
	Resource   string `json:"resource"`
	WorkflowID int    `json:"workflow_id"`
}

