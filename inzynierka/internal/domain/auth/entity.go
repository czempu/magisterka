package auth



type EffectivePermission struct {
	Resource  string
	Action    string
	Scope     string            // own|team|org
	Condition map[string]any    // np. {"max_days_back":7, "forbid_if_approved":true}
	Description *string 
}

// User to nasza encja użytkownika
type User struct {
    ID       int
    Username string
    Password string // w prodzie trzymaj HASH, nie plaintext!
}


type PermCtx struct {
	Scope     string
	Condition map[string]any
}

type UserScope struct {
	ID int
	Scope *string
}
