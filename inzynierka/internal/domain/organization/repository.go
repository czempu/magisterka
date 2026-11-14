package organization

import "context"

type Repository interface {
	GetOrgUnits(ctx context.Context) ([]OrgUnit, error)
	AddOrgUnit(ctx context.Context, name, code string, parentID *int) (string, error)
	AssignUser(ctx context.Context, userID, orgID int, title string, isPrimary bool) (string, error)
	SetManager(ctx context.Context, orgID, userID int, isDefault bool) (string, error)
	GetManagerForUser(ctx context.Context, userID int) (*int, error)
	    UpdateOrgUnit(ctx context.Context, id int, name, code string, parentID *int) (string, error)
    GetManagerForUnit(ctx context.Context, orgID int) (*int, error)
	GetOrgUnitByName(ctx context.Context, name string) (OrgUnit, error)
IsUserNowInUnit(ctx context.Context, userID, unitID int) (bool, error) 
DeleteOrg(ctx context.Context, orgID int) ( error)
    ListUserMemberships(ctx context.Context) ([]OrgMember, error)
}

