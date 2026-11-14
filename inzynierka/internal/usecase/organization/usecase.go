package organization

import (
	"context"
	do "erpbadger/internal/domain/organization"
)

type Usecase interface {
	GetOrgUnits(ctx context.Context) ([]do.OrgUnit, error)
	AddOrgUnit(ctx context.Context, name, code string, parentID *int) (string, error)
	AssignUser(ctx context.Context, userID, orgID int, title string, isPrimary bool) (string, error)
	SetManager(ctx context.Context, orgID, userID int, isDefault bool) (string, error)
	GetManagerForUser(ctx context.Context, userID int) (*int, error)
	    UpdateOrgUnit(ctx context.Context, id int, name, code string, parentID *int) (string, error)
    GetManagerForUnit(ctx context.Context, orgID int) (*int, error)
DeleteOrg(ctx context.Context, id int) error 

    // NOWE:
    ListUserMemberships(ctx context.Context) ([]do.OrgMember, error)
}

type usecase struct {
	repo do.Repository
}

func NewUsecase(r do.Repository) Usecase {
	return &usecase{repo: r}
}

func (u *usecase) DeleteOrg(ctx context.Context, id int) error {
	return u.repo.DeleteOrg(ctx, id)
}


func (u *usecase) GetOrgUnits(ctx context.Context) ([]do.OrgUnit, error) {
	return u.repo.GetOrgUnits(ctx)
}

func (u *usecase) AddOrgUnit(ctx context.Context, name, code string, parentID *int) (string, error) {
	return u.repo.AddOrgUnit(ctx, name, code, parentID)
}

func (u *usecase) AssignUser(ctx context.Context, userID, orgID int, title string, isPrimary bool) (string, error) {
	return u.repo.AssignUser(ctx, userID, orgID, title, isPrimary)
}

func (u *usecase) SetManager(ctx context.Context, orgID, userID int, isDefault bool) (string, error) {
	return u.repo.SetManager(ctx, orgID, userID, isDefault)
}

func (u *usecase) GetManagerForUser(ctx context.Context, userID int) (*int, error) {
	return u.repo.GetManagerForUser(ctx, userID)
}
func (u *usecase) UpdateOrgUnit(ctx context.Context, id int, name, code string, parentID *int) (string, error) {
    return u.repo.UpdateOrgUnit(ctx, id, name, code, parentID)
}

func (u *usecase) GetManagerForUnit(ctx context.Context, orgID int) (*int, error) {
    return u.repo.GetManagerForUnit(ctx, orgID)
}

 
func (u *usecase) ListUserMemberships(ctx context.Context) ([]do.OrgMember, error) {
    return u.repo.ListUserMemberships(ctx)
}

