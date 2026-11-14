package leaves

import "erpbadger/internal/domain/auth"

type Request struct {
	ID          int  
	Username  string
	Status string
	TypeCode string
	CreatedAt string
	Type *string
}

type LeaveTypes struct {
	Code string
	Name string
	Affects_balance int
	Count_strategy string
	Allow_days int
	Allow_hours int
	Allow_minutes int
	Min_inc_days int 
	Min_inc_hours int 
	Min_inc_minutes int 
	Nominal_day_minutes int
	Approval_policies_id int
}







type UserFilter struct {
    ID        *int    // Wskaźnik, bo 0 może być poprawnym ID lub wartością domyślną
    UserScope auth.UserScope 
}

type RequestInfo struct {
ID int
Request_id int
Start_at string
End_at string
Unit string
Quantity float64
Effective_minutes int
Note string
}


// NOWA STRUKTURA
type LeaveBalance struct {
    ID                     int    `json:"id"`
    EmployeeID             int    `json:"employee_id"`
    LeaveType              string `json:"leave_type"`
    AccrualDate            string `json:"accrual_date"`             // YYYY-MM-DD
    ExpirationDate         string `json:"expiration_date"`          // YYYY-MM-DD
    InitialBalanceMinutes  int    `json:"initial_balance_minutes"`
    RemainingBalanceMinutes int   `json:"remaining_balance_minutes"`
}
