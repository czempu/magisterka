package stats

import (
    "context"
	"fmt"
    do "erpbadger/internal/domain/stats"
)

var p = fmt.Print

type Usecase interface {

	GetEmpStats(ctx context.Context) (do.EmpStats,error)

}

type usecase struct {
    repo do.Repository
}

func NewUsecase(r do.Repository) Usecase {
    return &usecase{repo: r}
}


func (u *usecase) GetEmpStats(ctx context.Context) (do.EmpStats, error) {
return u.repo.GetEmpStats(ctx)
}














