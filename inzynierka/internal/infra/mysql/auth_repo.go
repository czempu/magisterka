package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"erpbadger/config"
	"erpbadger/internal/domain/auth"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) auth.Repository {
	return &AuthRepo{db: db}
}
func (r *AuthRepo) IsUserInSamePrimaryOrg(ctx context.Context, observerID, targetID int) (bool, error) {
	const q = `
		SELECT 1 
		FROM user_org_memberships 
		WHERE org_unit_id IN (
			SELECT org_unit_id 
			FROM user_org_memberships 
			WHERE is_primary = 1 AND user_id = ?
		) 
		AND user_id = ?
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRowContext(ctx, q, observerID, targetID).Scan(&exists)
	
	if err == sql.ErrNoRows {
		// Brak wyników oznacza, że target nie jest w dziale observera
		return false, nil
	}
	if err != nil {
		// Błąd bazy danych
		return false, err
	}

	// Znaleziono wiersz -> true
	return true, nil
}
// ForUser ładuje uprawnienia z widoku v_user_permissions (MySQL 8).
func strPtr(s string) *string { return &s }

func (r *AuthRepo) ForUser(ctx context.Context, userID int) ([]auth.EffectivePermission, error) {
	const q = `
		SELECT resource, action, scope, '' as conditions,case when x.resource = '*' then 'Całkowity dostęp' else m.description end as description 
		FROM v_user_permissions x
		left join module m on x.resource = m.module_name
		WHERE user_id = ?
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		if config.DevMode {
			return []auth.EffectivePermission{
				{Resource: "*", Action: "*", Scope: "org", Description: strPtr("Całkowity dostęp")},
				{Resource: "dashboard", Action: "read", Scope: "own", Description: strPtr("Panel Główny")},
				{Resource: "schedule", Action: "read", Scope: "own", Description: strPtr("Grafik")},
				{Resource: "leaves", Action: "read", Scope: "own", Description: strPtr("Urlopy")},
				{Resource: "users", Action: "read", Scope: "own", Description: strPtr("Użytkownicy")},
				{Resource: "permissions", Action: "read", Scope: "org", Description: strPtr("Uprawnienia")},
				{Resource: "organization", Action: "read", Scope: "org", Description: strPtr("Organizacja")},
				{Resource: "approvals", Action: "read", Scope: "own", Description: strPtr("Akceptacje")},
				{Resource: "timepolicy", Action: "read", Scope: "org", Description: strPtr("Polityka Czasu")},
				{Resource: "sick", Action: "read", Scope: "own", Description: strPtr("Choroby")},
				{Resource: "workflow", Action: "read", Scope: "own", Description: strPtr("Obiegi")},
			}, nil
		}
		return nil, err
	}
	defer rows.Close()

	perms := make([]auth.EffectivePermission, 0, 16)

	for rows.Next() {
		var ep auth.EffectivePermission
		var cond sql.NullString // JSON w MySQL zwracamy jako string

		if err := rows.Scan(&ep.Resource, &ep.Action, &ep.Scope, &cond, &ep.Description); err != nil {
			return nil, err
		}
		if cond.Valid && cond.String != "" {
			_ = json.Unmarshal([]byte(cond.String), &ep.Condition)
		}
		perms = append(perms, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}


	return perms, nil
}


func mergeRestrictivePerms(perms []auth.EffectivePermission) []auth.EffectivePermission {
	type key struct{ r, a, s string }
	m := map[key]auth.EffectivePermission{}
	for _, p := range perms {
		k := key{p.Resource, p.Action, p.Scope}
		if ex, ok := m[k]; ok {
			p.Condition = mergeRestrictive(ex.Condition, p.Condition)
		}
		m[k] = p
	}
	out := make([]auth.EffectivePermission, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func mergeRestrictive(a, b map[string]any) map[string]any {
	if a == nil && b == nil {
		return nil
	}
	r := map[string]any{}
	for k, v := range a {
		r[k] = v
	}
	for k, v := range b {
		switch k {
		case "max_days_back":
			r[k] = minInt(intFromAny(r[k]), intFromAny(v))
		case "forbid_if_approved", "workdays_only", "soft_correction":
			r[k] = truthy(r[k]) || truthy(v)
		default:
			r[k] = v
		}
	}
	return r
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if i, err := strconv.Atoi(t); err == nil {
			return i
		}
	}
	return 0
}

func minInt(a, b int) int {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

func truthy(v any) bool {
	b, _ := v.(bool)
	return b
}

// sanity check
func (r *AuthRepo) MustReady() {
	if r.db == nil {
		panic(errors.New("mysqlAuthRepo.db is nil"))
	}
}

func (r *AuthRepo) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
    var u auth.User
    query := "SELECT id, username, password FROM users WHERE username = ?"
    if err := r.db.QueryRowContext(ctx, query, username).
        Scan(&u.ID, &u.Username, &u.Password); err != nil {
        if config.DevMode {
            if username == "admin" {
                return &auth.User{ID: 1, Username: "admin", Password: "$2a$10$tkmDbMd.VL.9A6H0uA9Juurxe4m2Bj7neXiHpnwDL787/YqntUK/G"}, nil
            }
            if username == "user" {
                return &auth.User{ID: 2, Username: "user", Password: "$2a$10$3gE16ARJyCpBIIxGFKn3de.awo8N3haNLTPcmZUlmrnXcaXkbb0OW"}, nil
            }
        }
        return nil, err
    }
    return &u, nil
}

func (r *AuthRepo) GetById(ctx context.Context, id int) (*auth.User, error) {
    var u auth.User
    query := "SELECT id, username, password FROM users WHERE id = ?"
    if err := r.db.QueryRowContext(ctx, query, id).
        Scan(&u.ID, &u.Username, &u.Password); err != nil {
        if config.DevMode {
            if id == 1 {
                return &auth.User{ID: 1, Username: "admin", Password: "$2a$10$tkmDbMd.VL.9A6H0uA9Juurxe4m2Bj7neXiHpnwDL787/YqntUK/G"}, nil
            }
            if id == 2 {
                return &auth.User{ID: 2, Username: "user", Password: "$2a$10$3gE16ARJyCpBIIxGFKn3de.awo8N3haNLTPcmZUlmrnXcaXkbb0OW"}, nil
            }
        }
        return nil, err
    }
    return &u, nil
}


