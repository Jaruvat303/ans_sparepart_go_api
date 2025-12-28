package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(password, hashed string) error
}

type BcryptHasher struct {
	Cost int
}

func NewBcrypt(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{Cost: cost}
}

// hashPassword generates a bcrypt hash of the password.
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.Cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CompareHashAndPassword compares a password with a bcrypt hash.
func (b *BcryptHasher) CompareHashAndPassword(password, hashed string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))

	// Check err
	if err != nil {
		return err
	}
	return nil
}
