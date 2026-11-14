package users

//import "time"

type AddUser struct{
	Username string
	Email string
}

type DeleteUser struct{
	UserID int
}

// ModuleInfo to DTO dla modułów dostępnych danemu użytkownikowi
type UsersInfo struct {
    ID          int    `json:"id"`
    Username  string `json:"username"`
    Name string
    Surname string
}

type UserDetails struct {
	ID int
	Username string
	Email string
	Name string
	Surname string
}

type PII struct { // Personal identifying Information 
	ID int
	Name string
	Surname string
}

type UserPassChange struct {
	OldPassword string
	NewPassword string
}

