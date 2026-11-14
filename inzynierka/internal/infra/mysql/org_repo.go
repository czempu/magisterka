package mysql

import (
	"context"
	"database/sql"
	"erpbadger/internal/domain/organization"
	"fmt"
)

type OrgRepo struct {
	db *sql.DB
}

func NewOrgRepo(db *sql.DB) organization.Repository {
	return &OrgRepo{db: db}
}





func (r *OrgRepo) DeleteOrg(ctx context.Context, orgID int) (error){



  query4 := "delete from user_org_memberships uom where uom.org_unit_id = ?"
    _, err4 := r.db.ExecContext(ctx, query4, orgID)
    if err4 != nil {   
    	return err4
    }





  query := "delete from org_unit_managers where org_unit_id =?"
    _, err := r.db.ExecContext(ctx, query, orgID)
    if err != nil {   
    	return err
    }

	
  query2 := "update org_units set parent_id = null where parent_id =?"
    _, err2 := r.db.ExecContext(ctx, query2, orgID)
    if err2 != nil {   
    	return err2
    }


	
  query3 := "delete from org_units where id =?"
    _, err3 := r.db.ExecContext(ctx, query3, orgID)
    if err3 != nil {   
    	return err3
    }



    return nil
}





func (r *OrgRepo) UpdateOrgUnit(ctx context.Context, id int, name, code string, parentID *int) (string, error) {
    _, err := r.db.ExecContext(ctx,
        `UPDATE org_units SET name = ?, code = ?, parent_id = ? WHERE id = ?`,
        name, code, parentID, id,
    )
    if err != nil {
        return "", err
    }
    return "zaktualizowano jednostkę organizacyjną", nil
}

func (r *OrgRepo) GetManagerForUnit(ctx context.Context, orgID int) (*int, error) {
    q := `
        SELECT user_id
        FROM org_unit_managers
        WHERE org_unit_id = ? AND is_default = TRUE
        ORDER BY since_date DESC
        LIMIT 1;
    `
    var managerID *int
    err := r.db.QueryRowContext(ctx, q, orgID).Scan(&managerID)
    if err != nil {
        return nil, err
    }
    return managerID, nil
}

func (r *OrgRepo) GetOrgUnits(ctx context.Context) ([]organization.OrgUnit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, code, parent_id FROM org_units`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []organization.OrgUnit
	for rows.Next() {
		var ou organization.OrgUnit
		if err := rows.Scan(&ou.ID, &ou.Name, &ou.Code, &ou.ParentID); err != nil {
			return nil, err
		}
		res = append(res, ou)
	}
	return res, rows.Err()
}
func (r *OrgRepo) GetOrgUnitByName(ctx context.Context, name string) (organization.OrgUnit, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, name, code, parent_id FROM org_units WHERE name LIKE ?`, name)
    if err != nil {
        return organization.OrgUnit{}, err // Zwracamy pustą strukturę zamiast nil
    }
    defer rows.Close()

    var res organization.OrgUnit
    found := false
    for rows.Next() {
        if err := rows.Scan(&res.ID, &res.Name, &res.Code, &res.ParentID); err != nil {
            return organization.OrgUnit{}, err // Tutaj też pusta struktura
        }
        found = true
    }

    if !found {
        return organization.OrgUnit{}, fmt.Errorf("nie znaleziono jednostki o nazwie %s", name)
    }

    return res, rows.Err()
}








func (r *OrgRepo) AddOrgUnit(ctx context.Context, name, code string, parentID *int) (string, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO org_units (name, code, parent_id) VALUES (?, ?, ?)`, name, code, parentID)
	if err != nil {
		return "", err
	}
	return "dodano jednostkę organizacyjną", nil
}

func (r *OrgRepo) AssignUser(ctx context.Context, userID, orgID int, title string, isPrimary bool) (string, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_org_memberships (user_id, org_unit_id, title, is_primary, since_date) 
VALUES (?, ?, ?, ?, CURDATE()) 
ON DUPLICATE KEY UPDATE 
    org_unit_id = VALUES(org_unit_id), 
    title = VALUES(title), 
    is_primary = VALUES(is_primary)`,
		userID, orgID, title, isPrimary,
	)
	if err != nil {
		return "", err
	}
	return "użytkownik przypisany do jednostki", nil
}


func (r *OrgRepo) SetManager(ctx context.Context, orgID, userID int, isDefault bool) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// jeśli ustawiamy menedżera domyślnego – zdejmij flagę z innych
	if isDefault {
		if _, err := tx.ExecContext(ctx,
			`delete from org_unit_managers  
             WHERE org_unit_id = ?`,
			orgID,
		); err != nil {
			return "", err
		}
	}

	// UPSERT na parze (org_unit_id, user_id)
	_, err = tx.ExecContext(ctx, `
        INSERT INTO org_unit_managers (org_unit_id, user_id, since_date, is_default)
        VALUES (?, ?, CURDATE(), ?)
        ON DUPLICATE KEY UPDATE 
            since_date = VALUES(since_date),
            is_default = VALUES(is_default)
    `,
		orgID, userID, isDefault,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return "menedżer ustawiony", nil
}


func (r *OrgRepo) GetManagerForUser(ctx context.Context, userID int) (*int, error) {
	q := `
	SELECT m.user_id
	FROM user_org_memberships u
	JOIN org_unit_managers m ON m.org_unit_id = u.org_unit_id
	WHERE u.user_id = ? AND m.is_default = TRUE
	LIMIT 1;`
	var managerID *int
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&managerID)
	if err != nil {
		return nil, err
	}
	return managerID, nil
}
func (r *OrgRepo) ListUserMemberships(ctx context.Context) ([]organization.OrgMember, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT user_id, org_unit_id, title, is_primary
        FROM user_org_memberships
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var res []organization.OrgMember
    for rows.Next() {
        var m organization.OrgMember
        if err := rows.Scan(&m.UserID, &m.OrgUnitID, &m.Title, &m.IsPrimary); err != nil {
            return nil, err
        }
        res = append(res, m)
    }
    return res, rows.Err()
}



func (r *OrgRepo) IsUserNowInUnit(ctx context.Context, userID, unitID int) (bool, error) {
   var is string
	
	// Używamy QueryRowContext, bo spodziewamy się maksymalnie jednego wiersza
	err := r.db.QueryRowContext(ctx, 
		`select '1' as "is" from user_org_memberships where user_id = ? and org_unit_id = ? and since_date < now() and (until_date < now() or until_date is null)`, 
		userID, unitID,
	).Scan(&is)

	if err != nil {
		// Jeśli nie znaleziono wiersza, Scan zwróci błąd sql.ErrNoRows
		if err == sql.ErrNoRows {
			return false, nil // Brak roli to nie błąd, po prostu zwracamy false
		}
		// Inny błąd (np. utrata połączenia z bazą) zwracamy wyżej
		return false, fmt.Errorf("błąd podczas sprawdzania przynależności użytkownika do działu %d: %w", unitID, err)
	}

	// Jeśli Scan się udał, oznacza to, że wiersz istnieje (have będzie miało wartość '1')
	return true, nil
}
