package mysql

import (
    "context"
    "database/sql"
	"erpbadger/config"
	"erpbadger/internal/domain/modules"
	"fmt"
)

type ModulesRepo struct {
    db *sql.DB
}

func NewModulesRepo(db *sql.DB) modules.Repository {
    return &ModulesRepo{db: db}
}




func (r *ModulesRepo) GetModulesByUserID(ctx context.Context,userID int) ([]modules.ModuleInfo, error) {
   

	q := `
        SELECT m.id, m.module_name, m.description, umr.role, m.icon
          FROM users u
     LEFT JOIN user_module_roles umr ON u.id = umr.user_id
     LEFT JOIN module m               ON umr.module_id = m.id
         WHERE u.id = ?
    `
    rows, err := r.db.QueryContext(ctx, q, userID)
    if err != nil {
        if config.DevMode {
            return []modules.ModuleInfo{
                {ID: 1, ModuleName: "dashboard", Description: "Panel Główny", Role: "admin", Icon: "/public/lib/img/home.svg"},
                {ID: 2, ModuleName: "schedule", Description: "Grafik", Role: "admin", Icon: "/public/lib/img/calendar-clock.svg"},
                {ID: 3, ModuleName: "leaves", Description: "Urlopy", Role: "admin", Icon: "/public/lib/img/document.svg"},
                {ID: 4, ModuleName: "approvals", Description: "Akceptacje", Role: "admin", Icon: "/public/lib/img/checkbox.svg"},
                {ID: 5, ModuleName: "organization", Description: "Organizacja", Role: "admin", Icon: "/public/lib/img/building.svg"},
                {ID: 6, ModuleName: "users", Description: "Użytkownicy", Role: "admin", Icon: "/public/lib/img/users.svg"},
                {ID: 7, ModuleName: "permissions", Description: "Uprawnienia", Role: "admin", Icon: "/public/lib/img/unlock.svg"},
            }, nil
        }
        return nil, err
    }
    defer rows.Close()

    var module []modules.ModuleInfo
    for rows.Next() {
        var mi modules.ModuleInfo /// <--- tu błąd
        if err := rows.Scan(&mi.ID, &mi.ModuleName, &mi.Description, &mi.Role, &mi.Icon); err != nil {
            return nil, err
        }
        module = append(module, mi)
    }
    return module, rows.Err()
}


func (r *ModulesRepo) GetMeModules(ctx context.Context,userID int) ([]modules.ModuleInfo, error) {
   
	fmt.Printf("GetMeModules userID - %d\n", userID)
	q := `
   SELECT m.id, m.module_name, m.description, 'test' AS role, m.icon
FROM module m
WHERE EXISTS (
  SELECT 1 FROM v_user_permissions p
  WHERE p.user_id = ? AND p.resource='*'
)

UNION

SELECT 
    d.id AS id,   -- numeracja od 1
    y.resource AS module_name,
    d.description AS description,
    'test' AS role, d.icon
FROM (
    SELECT DISTINCT p.resource
    FROM v_user_permissions p
    WHERE p.user_id = ? AND p.resource <> '*'
) y
left join module d on y.resource = d.module_name 
order by id asc`
    rows, err := r.db.QueryContext(ctx, q, userID, userID)
    if err != nil {
        if config.DevMode {
            return []modules.ModuleInfo{
                {ID: 1, ModuleName: "dashboard", Description: "Panel Główny", Role: "admin", Icon: "/public/lib/img/home.svg"},
                {ID: 2, ModuleName: "schedule", Description: "Grafik", Role: "admin", Icon: "/public/lib/img/calendar-clock.svg"},
                {ID: 3, ModuleName: "leaves", Description: "Urlopy", Role: "admin", Icon: "/public/lib/img/document.svg"},
                {ID: 4, ModuleName: "approvals", Description: "Akceptacje", Role: "admin", Icon: "/public/lib/img/checkbox.svg"},
                {ID: 5, ModuleName: "organization", Description: "Organizacja", Role: "admin", Icon: "/public/lib/img/building.svg"},
                {ID: 6, ModuleName: "users", Description: "Użytkownicy", Role: "admin", Icon: "/public/lib/img/users.svg"},
                {ID: 7, ModuleName: "permissions", Description: "Uprawnienia", Role: "admin", Icon: "/public/lib/img/unlock.svg"},
            }, nil
        }
        return nil, err
    }
    defer rows.Close()

    var module []modules.ModuleInfo
    for rows.Next() {
        var mi modules.ModuleInfo
        if err := rows.Scan(&mi.ID, &mi.ModuleName, &mi.Description, &mi.Role, &mi.Icon); err != nil {
            return nil, err
        }
        module = append(module, mi)
    }
    return module, rows.Err()
}



func (r *ModulesRepo) GetSystemModules(ctx context.Context) ([]modules.SystemModules, error) {
   
	q := `
        SELECT * from module   `
    rows, err := r.db.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var module []modules.SystemModules
    for rows.Next() {
        var mi modules.SystemModules
        if err := rows.Scan(&mi.ID, &mi.ModuleName, &mi.Description, &mi.Icon); err != nil {
            return nil, err
        }
        module = append(module, mi)
    }
    return module, rows.Err()
}




