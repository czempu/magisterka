package modules

import "context"

type Repository interface {
	 GetSystemModules(ctx context.Context) ([]SystemModules, error)
    GetModulesByUserID(ctx context.Context ,  userID int) ([]ModuleInfo, error)
 GetMeModules(ctx context.Context, userID int) ([]ModuleInfo, error)
}
