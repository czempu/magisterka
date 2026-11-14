package employment

import (
	"fmt"
	"context"
	du "erpbadger/internal/domain/employment"
)

type Usecase interface {
	GetTerms(ctx context.Context, userID int) ([]du.EmploymentTerm, error)
	AddTerm(ctx context.Context, term du.EmploymentTermAdd) error
	UpdateTerm(ctx context.Context, term du.EmploymentTerm) error
DeleteTerm(ctx context.Context, id int) error
}

type usecase struct {
	repo du.Repository
}

func NewUsecase(r du.Repository) Usecase { return &usecase{repo: r} }

func (u *usecase) GetTerms(ctx context.Context, userID int) ([]du.EmploymentTerm, error) {
	return u.repo.GetEmploymentTerms(ctx, userID)
}
func (u *usecase) AddTerm(ctx context.Context, term du.EmploymentTermAdd) error {
    exist, err := u.repo.EmploymentTermExist(ctx, term.UserID, nil, term.EffectiveFrom, term.EffectiveTo)
    if err != nil {
        return fmt.Errorf("błąd podczas sprawdzania kolizji: %w", err)
    }
    if exist {
        return fmt.Errorf("istnieje kontrakt w tym przedziale")
    }
    return u.repo.AddEmploymentTerm(ctx, term)
}

func (u *usecase) UpdateTerm(ctx context.Context, term du.EmploymentTerm) error {
    excludeID := term.ID

    exist, err := u.repo.EmploymentTermExist(ctx, term.UserID, &excludeID, term.EffectiveFrom, term.EffectiveTo)
    if err != nil {
        return fmt.Errorf("błąd podczas sprawdzania kolizji: %w", err)
    }
    if exist {
        return fmt.Errorf("istnieje kontrakt w tym przedziale")
    }
    return u.repo.UpdateEmploymentTerm(ctx, term)
}
func (u *usecase) DeleteTerm(ctx context.Context, id int) error {
	return u.repo.DeleteEmploymentTerm(ctx, id)
}
