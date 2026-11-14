package sickness

import "context"

type Repository interface {
	List(ctx context.Context, employeeID *int, status *string) ([]Sickness, error)
	GetByID(ctx context.Context, id int) (*Sickness, error)
	Create(ctx context.Context, s Sickness) (int, error)
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status string) error
	SickExist(ctx context.Context, s Sickness) (bool,error)
}

