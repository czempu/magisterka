package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"erpbadger/config"
	"erpbadger/internal/domain/users"
)

type UsersRepo struct {
    db *sql.DB
}

func NewUsersRepo(db *sql.DB) users.Repository {
    return &UsersRepo{db: db}
}

func (r *UsersRepo) UserExists(ctx context.Context, username, email string, id *int) (bool, error) {
	if username == "" && email == "" {
		return false, errors.New("username and email are empty")
	}



if id == nil {

	fmt.Println("id jest nil")
	const q = `
SELECT EXISTS(
  SELECT 1
  FROM users
  WHERE (username = ? AND ? <> '')
     OR (email    = ? AND ? <> '')
) AS user_exists;
`

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, username, username, email, email).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

	fmt.Println(*id)
	const q = `
SELECT EXISTS(
  SELECT 1
  FROM users
  WHERE (username = ? AND ? <> '')
     and id <> ? 
     OR (email    = ? AND ? <> '')
     and id <> ? 
) AS user_exists;
`

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, username, username, *id, email, email, *id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil

}



func (r *UsersRepo) GetUserPII(ctx context.Context, userID int) (users.PII, error) {
	// Zapytanie SQL z dwoma podzapytaniami, zwracające jeden wiersz z dwiema kolumnami:
	// 'active_entries' (z timesheet_entries) i 'all_users' (z users).
	q := `
select id,name,surname from users where id = ?
	`

	var results users.PII

	err := r.db.QueryRowContext(ctx, q, userID).Scan(
		&results.ID,
		&results.Name, // Mapowanie 'active_entries' do pola Active
		&results.Surname,  // Mapowanie 'all_users' do pola Total
	)

	if err != nil {
		if config.DevMode {
			if userID == 1 {
				return users.PII{ID: 1, Name: "Administrator", Surname: "Systemu"}, nil
			}
			return users.PII{ID: 2, Name: "Jan", Surname: "Kowalski"}, nil
		}
		if err == sql.ErrNoRows {
			return users.PII{}, nil
		}
		return users.PII{}, err
	}
	return results, nil
}

func (r *UsersRepo) AddUser(ctx context.Context, username, email, passwordHash string) error {
	query := "insert into users (username, email, password) values (?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, username, email, passwordHash)
	return err
}

func (r *UsersRepo) PasswordChange(ctx context.Context, id int, passwordHash string) error {
	query := "update users set password = ? where id = ?"
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
} 



























func (r *UsersRepo) DeleteUser(ctx context.Context, userid int) (string, error) {
    query := "delete from users where id =?"
    _, err := r.db.ExecContext(ctx, query, userid)
    if err != nil {   
    	return "", err
    }
	info := "użytkownika usunięto."
    return info, nil
} 


func (r *UsersRepo) GetUsersInfo(ctx context.Context, scope string, userid int) ([]users.UsersInfo, error) {
	q := `
        select id,username, name, surname 
from users u where 1=1     `

	switch scope {

		case "own":
			q = q + fmt.Sprintf(" and id = %d", userid)

		case "team":
			q = q + fmt.Sprintf(" and u.id in (select user_id from user_org_memberships WHERE org_unit_id in (select org_unit_id from user_org_memberships where is_primary = 1 and user_id=%d))", userid)

		case "org":



		default:
			return nil, errors.New("brak uprawnień do modułu użytkowników (odczyt)")	
	}


    rows, err := r.db.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var user []users.UsersInfo
    for rows.Next() {
        var mi users.UsersInfo
        if err := rows.Scan(&mi.ID, &mi.Username, &mi.Name, &mi.Surname); err != nil {
            return nil, err
        }
        user = append(user, mi)
    }
    return user, rows.Err()
}


func (r *UsersRepo) UserDetails(ctx context.Context, userID int) (users.UserDetails, error) {
    q := `
        SELECT id, username, email,name,surname 
        FROM RCP.users WHERE id = ?;`
    
    var mi users.UserDetails
    err := r.db.QueryRowContext(ctx, q, userID).Scan(
        &mi.ID,
        &mi.Username,
        &mi.Email,
        &mi.Name,
        &mi.Surname,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return mi, nil         }
        return mi, err
    }

    return mi, nil
}





func (r *UsersRepo) UserDetailsEdit(ctx context.Context, userID int, userd users.UserDetails) (error) {


    q := `
        UPDATE RCP.users 
        SET 
            username = ?, 
            email = ?,
	    name = ?,
	    surname = ?,
        	updated_at = NOW()
         WHERE id = ?;`
    
    // Wykonanie zapytania UPDATE
    _, err := r.db.ExecContext(ctx, q, 
         userd.Username,userd.Email,userd.Name,userd.Surname,userID,
    )
    
    if err != nil {
        return  err
    }

    // Zwracamy zaktualizowane dane użytkownika
    return nil
}









