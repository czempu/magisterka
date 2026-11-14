package security

// PasswordHasher defines methods for securely hashing and verifying passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// PasswordGenerator defines methods for generating secure random passwords.
type PasswordGenerator interface {
	Generate(length int) (string, error)
}
