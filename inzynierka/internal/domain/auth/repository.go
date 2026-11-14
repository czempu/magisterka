package auth

import "context"



// Repository ukrywa szczegóły, skąd bierzemy listę IP
type Repository interface {
    ForUser(ctx context.Context, userID int) ([]EffectivePermission, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetById(ctx context.Context, id int) (*User, error)
IsUserInSamePrimaryOrg(ctx context.Context, observerID, targetID int) (bool, error)
    }

