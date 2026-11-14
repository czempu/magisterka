package leaves

import (
    "context"
	"fmt"
    du "erpbadger/internal/domain/leaves"
	apuc "erpbadger/internal/usecase/approval"
	auth "erpbadger/internal/usecase/auth"

)

var p = fmt.Print


type Usecase interface {

	AddRequest(ctx context.Context, iduser int, id,reason,start,end,unit string, quantity int, note string) (string, error) // ok nie trzeba bo zawsze użytkownik sam wypisuje urop 
	GetRequest(ctx context.Context,limit int , sort string, filter du.UserFilter) ([]du.Request, error) //ok read leaves
	GetRequestInfo(ctx context.Context,id int) ([]du.RequestInfo, error) //ok
	GetLeaveTypes(ctx context.Context) ([]du.LeaveTypes, error) // ok
    ListLeaveBalances(ctx context.Context, employeeID *int, leaveType *string) ([]du.LeaveBalance, error) //ok read leavebalance
    CreateLeaveBalance(ctx context.Context, b du.LeaveBalance) (int, error) //ok create leavebalance
    UpdateLeaveBalance(ctx context.Context, b du.LeaveBalance) error //ok
    DeleteLeaveBalance(ctx context.Context, id int) error




}

type usecase struct {
    repo du.Repository
	approvalSvc apuc.Usecase
	authSvc auth.Usecase

}


func NewUsecase(r du.Repository, ap apuc.Usecase, au auth.Usecase) Usecase {
	return &usecase{repo: r, approvalSvc: ap, authSvc: au}
}

func (u *usecase) GetRequest(ctx context.Context, limit int, sort string, filter du.UserFilter) ([]du.Request, error) {

	scope,userid, err := u.authSvc.MaxScopeFor(ctx, "leaves:read")
	if err != nil {
		return nil,fmt.Errorf("Brak Uprawnień")
	}
	filter.UserScope.Scope = &scope
	filter.UserScope.ID = userid

    return u.repo.GetRequest(ctx, limit, sort, filter)
}

func (u *usecase) GetRequestInfo(ctx context.Context, id int) ([]du.RequestInfo, error) {


var filter du.UserFilter
	scope,userid, err := u.authSvc.MaxScopeFor(ctx, "leaves:read")
	if err != nil {
		return nil,fmt.Errorf("Brak Uprawnień")
	}

filter.UserScope.Scope = &scope
filter.UserScope.ID = userid

    return u.repo.GetRequestInfo(ctx, id, filter)
}


//brak potrzeby uprawnień
func (u *usecase) GetLeaveTypes(ctx context.Context) ([]du.LeaveTypes, error) {
    return u.repo.GetLeaveTypes(ctx)
}


func (u *usecase) AddRequest(ctx context.Context, iduser int, id, reason, start, end, unit string, quantity int, note string) (string, error) {
   
	exist, errex := u.repo.LeaveExists(ctx,iduser, start,end)
    if errex != nil {
        return "", errex
    }
    if exist {
        return "", fmt.Errorf("Istnieje Urlop kolidujący z twoim aktywnym / oczekującym wnioskiem lub chorobą, (wniosek nie został utworzony)")
    }

_, reqID, _, err := u.repo.AddRequest(ctx, iduser, id, reason, start, end, unit, quantity, note)
    if err != nil {
        return "", err
    }

    Resource, err := u.repo.GetApprovalResourceByRequestID(ctx, reqID)
    if err != nil {
        return "błąd get resource from request by id", err
    }

    Response, err := u.approvalSvc.CreateApproval(ctx, Resource, int64(reqID), iduser)
    if err != nil {
        return "błąd approval create", err
    }

    return Response, err
}

// ==========================
//   LEAVE BALANCE – USECASE
// ==========================

func (u *usecase) ListLeaveBalances(
    ctx context.Context,
    employeeID *int,
    leaveType *string,
) ([]du.LeaveBalance, error) {




var filter du.UserFilter
	scope,userid, err := u.authSvc.MaxScopeFor(ctx, "leavebalance:read")
	if err != nil {
		return nil,fmt.Errorf("Brak Uprawnień")
	}


filter.UserScope.Scope = &scope
filter.UserScope.ID = userid



    return u.repo.ListLeaveBalances(ctx, employeeID, leaveType,filter)
}

func (u *usecase) CreateLeaveBalance(ctx context.Context, b du.LeaveBalance) (int, error) {
    if b.EmployeeID == 0 {
        return 0, fmt.Errorf("employee_id jest wymagany")
    }
    if b.LeaveType == "" {
        return 0, fmt.Errorf("leave_type jest wymagany")
    }
    if b.AccrualDate == "" || b.ExpirationDate == "" {
        return 0, fmt.Errorf("accrual_date i expiration_date są wymagane")
    }
    if b.InitialBalanceMinutes < 0 || b.RemainingBalanceMinutes < 0 {
        return 0, fmt.Errorf("minuty nie mogą być ujemne")
    }

    // Można dołożyć: jeśli Remaining == 0 na starcie, ustaw go na Initial.
    if b.RemainingBalanceMinutes == 0 && b.InitialBalanceMinutes > 0 {
        b.RemainingBalanceMinutes = b.InitialBalanceMinutes
    }

	scope,_, err := u.authSvc.MaxScopeFor(ctx, "leavebalance:create")
	if err != nil {
		return 0,fmt.Errorf("Brak Uprawnień")
	}

    	switch scope {
		case "own":

			is,_ := u.authSvc.CheckIsYourself(ctx,b.EmployeeID)
			if is {
			return u.repo.CreateLeaveBalance(ctx, b)
			} else {
				return 0,fmt.Errorf("Brak Uprawnień")
			}

		case "team":
			is,_ := u.authSvc.CheckIsUserInMyDepartment(ctx,b.EmployeeID)
			if is {
			return u.repo.CreateLeaveBalance(ctx, b)
			} else {
				return 0,fmt.Errorf("Brak Uprawnień")
			}


		case "org":
			return u.repo.CreateLeaveBalance(ctx, b)

		default:
			return 0,fmt.Errorf("Brak Uprawnień")

	}
    }

func (u *usecase) UpdateLeaveBalance(ctx context.Context, b du.LeaveBalance) error {
    if b.ID == 0 {
        return fmt.Errorf("id jest wymagane do aktualizacji")
    }
    if b.EmployeeID == 0 {
        return fmt.Errorf("employee_id jest wymagany")
    }
    if b.LeaveType == "" {
        return fmt.Errorf("leave_type jest wymagany")
    }
    if b.AccrualDate == "" || b.ExpirationDate == "" {
        return fmt.Errorf("accrual_date i expiration_date są wymagane")
    }
    if b.InitialBalanceMinutes < 0 || b.RemainingBalanceMinutes < 0 {
        return fmt.Errorf("minuty nie mogą być ujemne")
    }



	scope,_, err := u.authSvc.MaxScopeFor(ctx, "leavebalance:create")
	if err != nil {
		return fmt.Errorf("Brak Uprawnień")
	}
	
    	switch scope {
		case "own":

			is,_ := u.authSvc.CheckIsYourself(ctx,b.EmployeeID)
			if is {
			return u.repo.UpdateLeaveBalance(ctx, b)
			} else {
				return fmt.Errorf("Brak Uprawnień")
			}

		case "team":
			is,_ := u.authSvc.CheckIsUserInMyDepartment(ctx,b.EmployeeID)
			if is {
			return u.repo.UpdateLeaveBalance(ctx, b)
			} else {
				return fmt.Errorf("Brak Uprawnień")
			}


		case "org":
			return u.repo.UpdateLeaveBalance(ctx, b)

		default:
			return fmt.Errorf("Brak Uprawnień")

	}


}

func (u *usecase) DeleteLeaveBalance(ctx context.Context, id int) error {



    if id == 0 {
        return fmt.Errorf("id jest wymagane do usunięcia")
    }

	empid,err := u.repo.GetUserIdFromLeaveBalance(ctx,id)



	scope,_, err := u.authSvc.MaxScopeFor(ctx, "leavebalance:create")
	if err != nil {
		return fmt.Errorf("Brak Uprawnień")
	}
	
    	switch scope {
		case "own":

			is,_ := u.authSvc.CheckIsYourself(ctx,empid)
			if is {

    return u.repo.DeleteLeaveBalance(ctx, id)
			} else {
				return fmt.Errorf("Brak Uprawnień")
			}

		case "team":
			is,_ := u.authSvc.CheckIsUserInMyDepartment(ctx,empid)
			if is {

    return u.repo.DeleteLeaveBalance(ctx, id)
			} else {
				return fmt.Errorf("Brak Uprawnień")
			}


		case "org":

    return u.repo.DeleteLeaveBalance(ctx, id)

		default:
			return fmt.Errorf("Brak Uprawnień")

	}


}




