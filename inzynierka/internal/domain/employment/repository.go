package employment

import "context"

type Repository interface {
	GetEmploymentTerms(ctx context.Context, userID int) ([]EmploymentTerm, error)
	AddEmploymentTerm(ctx context.Context, term EmploymentTermAdd) error
	UpdateEmploymentTerm(ctx context.Context, term EmploymentTerm) error
DeleteEmploymentTerm(ctx context.Context, id int) error
	EmploymentTermExist(ctx context.Context,userid int, id *int, from string, to *string) (bool,error)
}

