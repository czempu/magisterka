package stats 

import "context"

type Repository interface{

	GetEmpStats(ctx context.Context) (EmpStats, error)

}
