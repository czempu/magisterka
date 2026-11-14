package mysql

import (
    "context"
    "database/sql"
    "erpbadger/internal/domain/leaves"
    "fmt"
    "errors"
)

type LeavesRepo struct {
    db *sql.DB
}

func NewLeavesRepo(db *sql.DB) leaves.Repository {
    return &LeavesRepo{db: db}
}


func (r *LeavesRepo) ListLeaveBalances(
    ctx context.Context,
    employeeID *int,
    leaveType *string, filter leaves.UserFilter,
) ([]leaves.LeaveBalance, error) {

    q := `
        SELECT 
            id,
            employee_id,
            leave_type,
            accrual_date,
            expiration_date,
            initial_balance_minutes,
            remaining_balance_minutes
        FROM leave_balance
        WHERE 1=1
    `
    args := []any{}

    if employeeID != nil {
        q += " AND employee_id = ?"
        args = append(args, *employeeID)
    }
    if leaveType != nil && *leaveType != "" {
        q += " AND leave_type = ?"
        args = append(args, *leaveType)
    }


if filter.UserScope.Scope != nil {
switch *filter.UserScope.Scope {

	case "org":
	//ok

	case "team":
		String := fmt.Sprintf("select user_id from user_org_memberships WHERE org_unit_id in (select org_unit_id from user_org_memberships where is_primary = 1 and user_id=%d)", filter.UserScope.ID)
	q = q + fmt.Sprintf(" and  employee_id in (%s)", String)	
	case "own":
	q = q + fmt.Sprintf(" and  employee_id = %d", filter.UserScope.ID)	
	default:
		return nil, errors.New("brak uprawnień")
}







}









    q += " ORDER BY employee_id, leave_type, accrual_date"

	fmt.Println(q)

    rows, err := r.db.QueryContext(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var res []leaves.LeaveBalance
    for rows.Next() {
        var b leaves.LeaveBalance
        if err := rows.Scan(
            &b.ID,
            &b.EmployeeID,
            &b.LeaveType,
            &b.AccrualDate,
            &b.ExpirationDate,
            &b.InitialBalanceMinutes,
            &b.RemainingBalanceMinutes,
        ); err != nil {
            return nil, err
        }
        res = append(res, b)
    }

    return res, rows.Err()
}
// ==========================
//   LEAVE BALANCE – INSERT
// ==========================

func (r *LeavesRepo) CreateLeaveBalance(ctx context.Context, b leaves.LeaveBalance) (int, error) {
    q := `
        INSERT INTO leave_balance (
            employee_id,
            leave_type,
            accrual_date,
            expiration_date,
            initial_balance_minutes,
            remaining_balance_minutes
        )
        VALUES (?, ?, ?, ?, ?, ?)
    `
    res, err := r.db.ExecContext(ctx, q,
        b.EmployeeID,
        b.LeaveType,
        b.AccrualDate,
        b.ExpirationDate,
        b.InitialBalanceMinutes,
        b.RemainingBalanceMinutes,
    )
    if err != nil {
        return 0, err
    }

    id, err := res.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("CreateLeaveBalance: LastInsertId: %w", err)
    }
    return int(id), nil
}

// ==========================
//   LEAVE BALANCE – UPDATE
// ==========================

func (r *LeavesRepo) UpdateLeaveBalance(ctx context.Context, b leaves.LeaveBalance) error {
    q := `
        UPDATE leave_balance
        SET
            employee_id = ?,
            leave_type = ?,
            accrual_date = ?,
            expiration_date = ?,
            initial_balance_minutes = ?,
            remaining_balance_minutes = ?
        WHERE id = ?
    `
    _, err := r.db.ExecContext(ctx, q,
        b.EmployeeID,
        b.LeaveType,
        b.AccrualDate,
        b.ExpirationDate,
        b.InitialBalanceMinutes,
        b.RemainingBalanceMinutes,
        b.ID,
    )
    return err
}

// ==========================
//   LEAVE BALANCE – DELETE
// ==========================

func (r *LeavesRepo) DeleteLeaveBalance(ctx context.Context, id int) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM leave_balance WHERE id = ?`, id)
    return err
}



func (r *LeavesRepo) AddRequest(
    ctx context.Context,
    iduser int,
    code, reason, start, end, unit string,
    quantity int,
    note string,
) (string, int, int, error) {

    // Ważne: zmienne sesyjne @... są per-connection, więc musimy użyć jednego połączenia.
    conn, err := r.db.Conn(ctx)
    if err != nil {
        return "", 0, 0, fmt.Errorf("db.Conn: %w", err)
    }
    defer conn.Close()

    query := `
CALL sp_create_leave_request_with_item(
    ?,               
    ?,
    ?, 
    STR_TO_DATE(NULLIF(?, ''), '%Y-%m-%dT%H:%i'),
    STR_TO_DATE(NULLIF(?, ''), '%Y-%m-%dT%H:%i'), 
    ?, 
    ?,            
    ?,                  
    @req_id,           
    @item_id          
);
`

    // Uwaga: Twoja procedura ma in_quantity jako DECIMAL(10,2),
    _, err = conn.ExecContext(ctx, query,
        iduser,
        code,
        reason,
        start,
        end,
        unit,
        quantity,
        note,
    )
    if err != nil {
        return "", 0, 0, err
    }

    // Pobierz OUT-y + dodatkowe @approval_id ustawiane w procedurze (sesyjnie).
    var reqID, itemID int
    var approvalID sql.NullInt64

    err = conn.QueryRowContext(ctx, "SELECT @req_id, @item_id, @approval_id").
        Scan(&reqID, &itemID, &approvalID)
    if err != nil {
        return "", 0, 0, fmt.Errorf("błąd pobierania ID: %w", err)
    }

    // Zachowujemy kompatybilność: nadal zwracamy reqID i itemID.
    // approvalID nie ma w kontrakcie — dokładamy do message.
    if approvalID.Valid {
        return fmt.Sprintf(
            "Dodano wniosek. req_id=%d, item_id=%d, approval_id=%d",
            reqID, itemID, approvalID.Int64,
        ), reqID, itemID, nil
    }

    return fmt.Sprintf(
        "Dodano wniosek. req_id=%d, item_id=%d",
        reqID, itemID,
    ), reqID, itemID, nil
}

func (r *LeavesRepo) GetRequest(ctx context.Context, limit int,sort string, filter leaves.UserFilter) ([]leaves.Request, error) {
	q := `
select 
  lr.id,username
  ,case when ar.status is null then 'brak danych' else ar.status end as status
  ,lr.type_code, lr.created_at ed_at, x.name
 from leave_requests lr 
left join users u on lr.employee_id = u.id   
left join leave_types x on x.code = lr.type_code
left join approval_policies ap on ap.id = x.approval_policies_id
left join approval_requests ar on ap.resource = ar.resource and ar.target_id = lr.id
where 1=1  
  `

  if filter.ID != nil {
	q = q + fmt.Sprintf(" and u.id = %d",*filter.ID)
  }

if filter.UserScope.Scope != nil {
switch *filter.UserScope.Scope {

	case "org":

	case "team":
		String := fmt.Sprintf("select user_id from user_org_memberships WHERE org_unit_id in (select org_unit_id from user_org_memberships where is_primary = 1 and user_id=%d)", filter.UserScope.ID)
	q = q + fmt.Sprintf(" and  u.id in (%s)", String)	
	case "own":
	q = q + fmt.Sprintf(" and  u.id = %d", filter.UserScope.ID)	
	default:
		return nil, errors.New("brak uprawnień")
}







}




if sort != ""  {
	if sort == "date" {
		q = q + " order by lr.created_at asc"
	}
	if sort == "-date" {
		q = q + " order by lr.created_at desc"
	}
}


if limit != 0 {
	q = q + fmt.Sprintf(" LIMIT %d", limit)	
}

fmt.Println(q)


    rows, err := r.db.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []leaves.Request
    for rows.Next() {
        var mi leaves.Request
        if err := rows.Scan(&mi.ID, &mi.Username, &mi.Status, &mi.TypeCode, &mi.CreatedAt, &mi.Type); err != nil {
            return nil, err
        }
        users = append(users, mi)
    }
    return users, rows.Err()
}

func (r *LeavesRepo) GetRequestInfo(ctx context.Context, id int,  filter leaves.UserFilter) ([]leaves.RequestInfo, error) {
	q := `
     select lri.* from leave_request_items lri 
left join leave_requests lr on lri.request_id = lr.id
where request_id = ? `


switch *filter.UserScope.Scope {

	case "org":
	//ok

	case "team":
		String := fmt.Sprintf("select user_id from user_org_memberships WHERE org_unit_id in (select org_unit_id from user_org_memberships where is_primary = 1 and user_id=%d)", filter.UserScope.ID)
	q = q + fmt.Sprintf(" and employee_id in (%s)", String)	
	case "own":
	q = q + fmt.Sprintf(" and employee_id = %d", filter.UserScope.ID)	
	default:
		return nil, errors.New("brak uprawnień")
}







    rows, err := r.db.QueryContext(ctx, q, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []leaves.RequestInfo
    for rows.Next() {
        var mi leaves.RequestInfo
        if err := rows.Scan(&mi.ID,
&mi.Request_id,
&mi.Start_at,
&mi.End_at,
&mi.Unit,
&mi.Quantity,
&mi.Effective_minutes,
&mi.Note); err != nil {
            return nil, err
        }
        users = append(users, mi)
    }
    return users, rows.Err()
}

func (r *LeavesRepo) GetLeaveTypes(ctx context.Context) ([]leaves.LeaveTypes, error) {
	q := `
     select * from leave_types `
    rows, err := r.db.QueryContext(ctx, q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []leaves.LeaveTypes
    for rows.Next() {
        var mi leaves.LeaveTypes
        if err := rows.Scan(&mi.Code,
&mi.Name,
&mi.Affects_balance,
&mi.Count_strategy,
&mi.Allow_days,
&mi.Allow_hours,
&mi.Allow_minutes,
&mi.Min_inc_days,
&mi.Min_inc_hours,
&mi.Min_inc_minutes,
&mi.Nominal_day_minutes,
&mi.Approval_policies_id); err != nil {
            return nil, err
        }
        users = append(users, mi)
    }
    return users, rows.Err()
}






func (r *LeavesRepo) GetApprovalResourceByRequestID(ctx context.Context, ID int) (string, error) {
	q := `
		select resource from approval_policies ap  
		where id = (select lt.approval_policies_id from leave_types 
		lt where lt.code = (select type_code from leave_requests where id = ?))
		`

	var Resource string
	err := r.db.QueryRowContext(ctx, q, ID).Scan(&Resource)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("nie znaleziono polityki dla request_id=%d", ID)
		}
		return "", fmt.Errorf("błąd podczas pobierania approval policy: %w", err)
	}

	return Resource, nil
}





func (r *LeavesRepo) GetUserIdFromLeaveBalance(ctx context.Context, id int) (int, error) {
	q := `	select lb.employee_id  from leave_balance lb where id = ? 
	`

	var eid int
	err := r.db.QueryRowContext(ctx, q, id).Scan(&eid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("nie znaleziono numeru pracownika dla bilansu o id =%d", id)
		}
		return 0, fmt.Errorf("błąd w pobieraniu id usera z bilansu: %w", err)
	}

	return eid, nil
}













 func (r *LeavesRepo)      LeaveExists(ctx context.Context, userid int,start,end string) (bool, error) {

	const q = `
        WITH tmp AS (
  SELECT start_at AS start_at, end_at AS end_at
  FROM leave_request_items
  WHERE request_id IN (
    SELECT target_id
    FROM approval_requests
    WHERE requester_id = ?
      AND resource = 'leave_requests'
      AND status IN ('approved', 'pending')
  )
),
tmp2 AS (
  SELECT from_date AS start_at, to_date AS end_at
  FROM employee_sickness
  WHERE id IN (
    SELECT target_id
    FROM approval_requests
    WHERE requester_id = ?
      AND resource = 'sick'
      AND status IN ('approved', 'rejected')
  )
),
tmp3 AS (
  SELECT start_at, end_at FROM tmp
  UNION ALL
  SELECT start_at, end_at FROM tmp2
)
SELECT EXISTS (
  SELECT 1
  FROM tmp3 t
  WHERE STR_TO_DATE(NULLIF(?, ''), '%Y-%m-%dT%H:%i') < t.end_at
    AND STR_TO_DATE(NULLIF(?, ''), '%Y-%m-%dT%H:%i') > t.start_at
) AS is_collision;`

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, userid, userid, start, end).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil



 }

