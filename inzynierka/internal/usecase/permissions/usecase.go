package permissions

import (
    "context"
    "errors"

    du "erpbadger/internal/domain/permissions"
)

type Usecase interface {


// uprawnienia tutaj na poziomie routera - permissions tylko read dla get, dla reszty create na poziomie org

    GetSystemRoles(ctx context.Context) ([]du.SystemRoles, error)
    GetRoleByUserID(ctx context.Context, userID int) ([]du.Role, error)
    UserRoleChange(ctx context.Context, userid, roleid int) (string, error)
	DoesUserHaveRole(ctx context.Context, userID, roleID int)(bool, error)

    GetRoles(ctx context.Context) ([]du.RolePure, error)
    GetRoleByID(ctx context.Context, id int) (*du.RolePure, error)
    CreateRole(ctx context.Context, slug, name string) (int, error)
    UpdateRole(ctx context.Context, id int, slug, name string) error
    DeleteRole(ctx context.Context, id int) error

    GetPermissions(ctx context.Context) ([]du.Permission, error)
    GetPermissionByID(ctx context.Context, id int) (*du.Permission, error)
    CreatePermission(ctx context.Context, p du.Permission) (int, error)
    UpdatePermission(ctx context.Context, p du.Permission) error
    DeletePermission(ctx context.Context, id int) error

    AssignPermissionToRole(ctx context.Context, roleID, permID int) error
    RemovePermissionFromRole(ctx context.Context, roleID, permID int) error

    AssignRoleToUser(ctx context.Context, userID, roleID int) error
    RemoveRoleFromUser(ctx context.Context, userID, roleID int) error

GetPermissionsAssignedToRole(ctx context.Context, roleID int) ([]int, error)

}

type usecase struct {
    repo du.Repository
}

func NewUsecase(r du.Repository) Usecase {
    return &usecase{repo: r}
}

//
// === STARE — BEZ ZMIAN ===
//

func (u *usecase) GetSystemRoles(ctx context.Context) ([]du.SystemRoles, error) {
    return u.repo.GetSystemRoles(ctx)
}
func (u *usecase) DoesUserHaveRole(ctx context.Context, userID,roleID int)(bool, error) {
	return u.repo.DoesUserHaveRole(ctx, userID,roleID)
}
func (u *usecase) GetRoleByUserID(ctx context.Context, userID int) ([]du.Role, error) {
    return u.repo.GetRoleByUserID(ctx, userID)
}

func (u *usecase) UserRoleChange(ctx context.Context, userid, roleid int) (string, error) {
    return u.repo.UserRoleChange(ctx, userid, roleid)
}

//
// === NOWE REST — ROLES ===
//
func (u *usecase) GetPermissionsAssignedToRole(ctx context.Context, roleID int) ([]int, error) {
    return u.repo.GetPermissionsAssignedToRole(ctx, roleID)
}

func (u *usecase) GetRoles(ctx context.Context) ([]du.RolePure, error) {
    return u.repo.GetRoles(ctx)
}

func (u *usecase) GetRoleByID(ctx context.Context, id int) (*du.RolePure, error) {
    return u.repo.GetRoleByID(ctx, id)
}


func (u *usecase) CreateRole(ctx context.Context, slug, name string) (int, error) {
    if slug == "" || name == "" {
        return 0, errors.New("slug and name are required")
    }
    return u.repo.CreateRole(ctx, slug, name)
}

func (u *usecase) UpdateRole(ctx context.Context, id int, slug, name string) error {
    if id <= 0 {
        return errors.New("invalid id")
    }
    return u.repo.UpdateRole(ctx, id, slug, name)
}

func (u *usecase) DeleteRole(ctx context.Context, id int) error {
    return u.repo.DeleteRole(ctx, id)
}

//
// === NOWE REST — PERMISSIONS ===
//

func (u *usecase) GetPermissions(ctx context.Context) ([]du.Permission, error) {
    return u.repo.GetPermissions(ctx)
}

func (u *usecase) GetPermissionByID(ctx context.Context, id int) (*du.Permission, error) {
    return u.repo.GetPermissionByID(ctx, id)
}

func (u *usecase) CreatePermission(ctx context.Context, p du.Permission) (int, error) {
    if p.Resource == "" || p.Action == "" || p.Scope == "" {
        return 0, errors.New("missing required fields")
    }
    return u.repo.CreatePermission(ctx, p)
}

func (u *usecase) UpdatePermission(ctx context.Context, p du.Permission) error {
    if p.ID <= 0 {
        return errors.New("invalid id")
    }
    return u.repo.UpdatePermission(ctx, p)
}

func (u *usecase) DeletePermission(ctx context.Context, id int) error {
    return u.repo.DeletePermission(ctx, id)
}

//
// === NOWE REST — RELATIONS ===
//

func (u *usecase) AssignPermissionToRole(ctx context.Context, roleID, permID int) error {
    return u.repo.AssignPermissionToRole(ctx, roleID, permID)
}

func (u *usecase) RemovePermissionFromRole(ctx context.Context, roleID, permID int) error {
    return u.repo.RemovePermissionFromRole(ctx, roleID, permID)
}

func (u *usecase) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
    return u.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (u *usecase) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
    return u.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

