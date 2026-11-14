package leaves

import "context"
type Repository interface {
	 GetRequest(ctx context.Context,limit int, sort string, filter UserFilter) ([]Request, error)
GetRequestInfo(ctx context.Context, id int, filter UserFilter) ([]RequestInfo, error)
GetLeaveTypes(ctx context.Context) ([]LeaveTypes, error)
AddRequest(ctx context.Context, iduser int,id,reason,start,end,unit string, quantity int, note string) (string,int,int, error)
GetApprovalResourceByRequestID(ctx context.Context, ID int) (string, error)
    ListLeaveBalances(ctx context.Context, employeeID *int, leaveType *string, filter UserFilter) ([]LeaveBalance, error)
    CreateLeaveBalance(ctx context.Context, b LeaveBalance) (int, error)
    UpdateLeaveBalance(ctx context.Context, b LeaveBalance) error
    DeleteLeaveBalance(ctx context.Context, id int) error
GetUserIdFromLeaveBalance(ctx context.Context, id int) (int, error)
LeaveExists(ctx context.Context,userid int , start,end string) (bool, error)

}
