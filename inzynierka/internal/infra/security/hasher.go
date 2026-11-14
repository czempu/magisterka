package security

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements security.PasswordHasher using bcrypt.
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// RandomPasswordGenerator implements security.PasswordGenerator using crypto/rand.
type RandomPasswordGenerator struct{}

func NewRandomPasswordGenerator() *RandomPasswordGenerator {
	return &RandomPasswordGenerator{}
}

func (g *RandomPasswordGenerator) Generate(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}
	return string(password), nil
}
