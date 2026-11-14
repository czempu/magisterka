
package permissions
type SystemRoles struct {
    ID   int    `json:"id"`
    Slug string `json:"slug"`
    Name string `json:"name"`
}

type Role struct {
    UserID int    `json:"userid"`
    RoleID int    `json:"roleid"`
    Slug   string `json:"slug"`
    Name   string `json:"name"`
}
type RolePure struct {
    ID   int    `json:"id"`
    Slug string `json:"slug"`
    Name string `json:"name"`
}

type PermissionConditions struct {
    MaxDaysBack      int  `json:"max_days_back"`
    WorkdaysOnly     bool `json:"workdays_only"`
    ForbidIfApproved bool `json:"forbid_if_approved"`
}

type Permission struct {
    ID         int                   `json:"id"`
    Resource   string                `json:"resource"`
    Action     string                `json:"action"`
    Scope      string                `json:"scope"`
    Conditions *PermissionConditions `json:"conditions"`
    Description *string 
}

type RolePermission struct {
    RoleID       int `json:"role_id"`
    PermissionID int `json:"permission_id"`
}

type UserRole struct {
    UserID int `json:"user_id"`
    RoleID int `json:"role_id"`
}
type RolePermissionAssignment struct {
    PermissionID int `json:"permission_id"`
}

