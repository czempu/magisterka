package users

import "context"

type Repository interface {
	AddUser(ctx context.Context, username, email, passwordHash string) error
	PasswordChange(ctx context.Context, id int, passwordHash string) error
	DeleteUser(ctx context.Context, userID int) (string, error)
	GetUsersInfo(ctx context.Context, scope string, userid int) ([]UsersInfo, error)
	UserDetails(ctx context.Context, userID int) (UserDetails, error)
	UserDetailsEdit(ctx context.Context, userID int, userd UserDetails) error
	GetUserPII(ctx context.Context, userID int) (PII, error)
	UserExists(ctx context.Context, username, email string, id *int) (bool, error)
}
