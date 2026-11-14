package modules

import (
    "context"
	"fmt"
    du "erpbadger/internal/domain/modules"
)

var p = fmt.Print



// Usecase – interfejs dla handlera/middleware
type Usecase interface {

 GetSystemModules(ctx context.Context) ([]du.SystemModules, error)
 GetMeModules(ctx context.Context, userID int) ([]du.ModuleInfo, error)
    GetUserModules(ctx context.Context, userID int) ([]du.ModuleInfo, error)
}


type usecase struct {
    repo du.Repository
}

// NewUsecase – wstrzykujesz tutaj repo (fake lub SQL)
func NewUsecase(r du.Repository) Usecase {
    return &usecase{repo: r}
}



func (u *usecase) GetSystemModules(ctx context.Context) ([]du.SystemModules, error) {
    return u.repo.GetSystemModules(ctx)
}

// GetUserModules – po prostu deleguje do repo
func (u *usecase) GetUserModules(ctx context.Context, userID int) ([]du.ModuleInfo, error) {
    return u.repo.GetModulesByUserID(ctx, userID)
}

// GetUserModules – po prostu deleguje do repo
func (u *usecase) GetMeModules(ctx context.Context, userID int) ([]du.ModuleInfo, error) {
    return u.repo.GetMeModules(ctx, userID)
}


























