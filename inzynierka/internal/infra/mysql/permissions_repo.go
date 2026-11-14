package mysql

import (
    "context"
    "database/sql"
    "encoding/json"
	"fmt"
    "erpbadger/internal/domain/permissions"
)

type PermissionsRepo struct {
    db *sql.DB
}

func NewPermissionsRepo(db *sql.DB) permissions.Repository {
    return &PermissionsRepo{db: db}
}

func (r *PermissionsRepo) GetSystemRoles(ctx context.Context) ([]permissions.SystemRoles, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, slug, name FROM roles`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []permissions.SystemRoles
    for rows.Next() {
        var x permissions.SystemRoles
        if err := rows.Scan(&x.ID, &x.Slug, &x.Name); err != nil {
            return nil, err
        }
        out = append(out, x)
    }
    return out, rows.Err()
}
func (r *PermissionsRepo) DoesUserHaveRole(ctx context.Context, userID, roleID int)(bool, error) {
var have string
	
	// Używamy QueryRowContext, bo spodziewamy się maksymalnie jednego wiersza
	err := r.db.QueryRowContext(ctx, 
		"SELECT '1' FROM user_roles WHERE user_id = ? AND role_id = ?", 
		userID, roleID,
	).Scan(&have)

	if err != nil {
		// Jeśli nie znaleziono wiersza, Scan zwróci błąd sql.ErrNoRows
		if err == sql.ErrNoRows {
			return false, nil // Brak roli to nie błąd, po prostu zwracamy false
		}
		// Inny błąd (np. utrata połączenia z bazą) zwracamy wyżej
		return false, fmt.Errorf("błąd podczas sprawdzania roli użytkownika: %w", err)
	}

	// Jeśli Scan się udał, oznacza to, że wiersz istnieje (have będzie miało wartość '1')
	return true, nil
}
func (r *PermissionsRepo) GetRoleByUserID(ctx context.Context, userID int) ([]permissions.Role, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT ur.user_id, ur.role_id, r.slug, r.name
         FROM user_roles ur
         LEFT JOIN roles r ON r.id = ur.role_id
         WHERE ur.user_id = ?`,
        userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []permissions.Role
    for rows.Next() {
        var x permissions.Role
        if err := rows.Scan(&x.UserID, &x.RoleID, &x.Slug, &x.Name); err != nil {
            return nil, err
        }
        out = append(out, x)
    }
    return out, rows.Err()
}

func (r *PermissionsRepo) UserRoleChange(ctx context.Context, uid, rid int) (string, error) {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO user_roles (user_id, role_id)
         VALUES (?,?)
         ON DUPLICATE KEY UPDATE role_id = VALUES(role_id)`,
        uid, rid)
    if err != nil {
        return "", err
    }
    return "zaktualizowano", nil
}

//
// ===== NOWY MODUŁ: ROLES CRUD =====
//
func (r *PermissionsRepo) GetPermissionsAssignedToRole(ctx context.Context, roleID int) ([]int, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT permission_id FROM role_permissions WHERE role_id = ?`,
        roleID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []int
    for rows.Next() {
        var id int
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        out = append(out, id)
    }
    return out, rows.Err()
}

func (r *PermissionsRepo) GetRoles(ctx context.Context) ([]permissions.RolePure, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, slug, name FROM roles ORDER BY id`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []permissions.RolePure
    for rows.Next() {
        var x permissions.RolePure
        if err := rows.Scan(&x.ID, &x.Slug, &x.Name); err != nil {
            return nil, err
        }
        out = append(out, x)
    }
    return out, rows.Err()
}

func (r *PermissionsRepo) GetRoleByID(ctx context.Context, id int) (*permissions.RolePure, error) {
    row := r.db.QueryRowContext(ctx,
        `SELECT id, slug, name FROM roles WHERE id = ?`,
        id)

    var x permissions.RolePure
    if err := row.Scan(&x.ID, &x.Slug, &x.Name); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &x, nil
}

func (r *PermissionsRepo) CreateRole(ctx context.Context, slug, name string) (int, error) {
    res, err := r.db.ExecContext(ctx,
        `INSERT INTO roles (slug, name) VALUES (?, ?)`,
        slug, name)
    if err != nil {
        return 0, err
    }
    id, _ := res.LastInsertId()
    return int(id), nil
}

func (r *PermissionsRepo) UpdateRole(ctx context.Context, id int, slug, name string) error {
    _, err := r.db.ExecContext(ctx,
        `UPDATE roles SET slug=?, name=? WHERE id=?`,
        slug, name, id)
    return err
}

func (r *PermissionsRepo) DeleteRole(ctx context.Context, id int) error {
    _, err := r.db.ExecContext(ctx,
        `DELETE FROM roles WHERE id=?`,
        id)
    return err
}

//
// ===== NEW: PERMISSIONS CRUD =====
//

func (r *PermissionsRepo) GetPermissions(ctx context.Context) ([]permissions.Permission, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT x.id, resource, action, scope, '' as conditions, case when x.resource = '*' then 'Całkowity dostęp' else m.description end as description 
         FROM permissions x
         	left join module m on x.resource = m.module_name 
         ORDER BY x.resource,x.action`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []permissions.Permission

    for rows.Next() {
        var x permissions.Permission
        var cond sql.NullString

        if err := rows.Scan(&x.ID, &x.Resource, &x.Action, &x.Scope, &cond, &x.Description); err != nil {
            return nil, err
        }

        if cond.Valid {
            var c permissions.PermissionConditions
            _ = json.Unmarshal([]byte(cond.String), &c)
            x.Conditions = &c
        }

        out = append(out, x)
    }

    return out, rows.Err()
}

func (r *PermissionsRepo) GetPermissionByID(ctx context.Context, id int) (*permissions.Permission, error) {
    var x permissions.Permission
    var cond sql.NullString

    err := r.db.QueryRowContext(ctx,
        `SELECT id, resource, action, scope, '' as conditions
         FROM permissions WHERE id=?`,
        id).Scan(&x.ID, &x.Resource, &x.Action, &x.Scope, &cond)

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    if cond.Valid {
        var c permissions.PermissionConditions
        _ = json.Unmarshal([]byte(cond.String), &c)
        x.Conditions = &c
    }

    return &x, nil
}

func (r *PermissionsRepo) CreatePermission(ctx context.Context, p permissions.Permission) (int, error) {
    var cond *string
    if p.Conditions != nil {
        b, _ := json.Marshal(p.Conditions)
        s := string(b)
        cond = &s
    }

    res, err := r.db.ExecContext(ctx,
        `INSERT INTO permissions (resource, action, scope, conditions) 
         VALUES (?,?,?,?)`,
        p.Resource, p.Action, p.Scope, cond)
    if err != nil {
        return 0, err
    }

    id, _ := res.LastInsertId()
    return int(id), nil
}

func (r *PermissionsRepo) UpdatePermission(ctx context.Context, p permissions.Permission) error {
    var cond *string
    if p.Conditions != nil {
        b, _ := json.Marshal(p.Conditions)
        s := string(b)
        cond = &s
    }

    _, err := r.db.ExecContext(ctx,
        `UPDATE permissions 
         SET resource=?, action=?, scope=?, conditions=? 
         WHERE id=?`,
        p.Resource, p.Action, p.Scope, cond, p.ID)

    return err
}

func (r *PermissionsRepo) DeletePermission(ctx context.Context, id int) error {
    _, err := r.db.ExecContext(ctx,
        `DELETE FROM permissions WHERE id=?`,
        id)
    return err
}

//
// ===== RELATIONS =====
//

func (r *PermissionsRepo) AssignPermissionToRole(ctx context.Context, roleID, permID int) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO role_permissions (role_id, permission_id)
         VALUES (?,?)
         ON DUPLICATE KEY UPDATE permission_id = permission_id`,
        roleID, permID)
    return err
}

func (r *PermissionsRepo) RemovePermissionFromRole(ctx context.Context, roleID, permID int) error {
    _, err := r.db.ExecContext(ctx,
        `DELETE FROM role_permissions WHERE role_id=? AND permission_id=?`,
        roleID, permID)
    return err
}

func (r *PermissionsRepo) AssignRoleToUser(ctx context.Context, userID, roleID int) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO user_roles (user_id, role_id)
         VALUES (?,?)
         ON DUPLICATE KEY UPDATE role_id = VALUES(role_id)`,
        userID, roleID)
    return err
}

func (r *PermissionsRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
    _, err := r.db.ExecContext(ctx,
        `DELETE FROM user_roles WHERE user_id=? AND role_id=?`,
        userID, roleID)
    return err
}




