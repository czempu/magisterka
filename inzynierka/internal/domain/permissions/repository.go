package permissions

import "context"

type Repository interface {
    GetSystemRoles(ctx context.Context) ([]SystemRoles, error)
    UserRoleChange(ctx context.Context, userid, roleid int) (string, error)
    GetRoleByUserID(ctx context.Context, userID int) ([]Role, error)
DoesUserHaveRole(ctx context.Context, userID, roleID int)(bool, error)
    GetRoles(ctx context.Context) ([]RolePure, error)
    GetRoleByID(ctx context.Context, id int) (*RolePure, error)
    CreateRole(ctx context.Context, slug, name string) (int, error)
    UpdateRole(ctx context.Context, id int, slug, name string) error
    DeleteRole(ctx context.Context, id int) error
    GetPermissions(ctx context.Context) ([]Permission, error)
    GetPermissionByID(ctx context.Context, id int) (*Permission, error)
    CreatePermission(ctx context.Context, p Permission) (int, error)
    UpdatePermission(ctx context.Context, p Permission) error
    DeletePermission(ctx context.Context, id int) error
    AssignPermissionToRole(ctx context.Context, roleID, permID int) error
    RemovePermissionFromRole(ctx context.Context, roleID, permID int) error
    AssignRoleToUser(ctx context.Context, userID, roleID int) error
    RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
GetPermissionsAssignedToRole(ctx context.Context, roleID int) ([]int, error)

}

