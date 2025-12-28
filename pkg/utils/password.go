package utils

import (
	"fmt"
	"regexp"
)

// Define regex for password strength validation
var (
	minLengthRegex  = regexp.MustCompile(`.{8,}`)
	hasNumberRegex  = regexp.MustCompile(`[0-9]`)
	hasUpperRegex   = regexp.MustCompile(`[A-Z]`)
	hasLowerRegex   = regexp.MustCompile(`[a-z]`)
	hasSpecialRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

// VerifyPasswordStrength verifies if a password meets the strength criteria.
func VerifyPasswordStrength(password string) error {
	if !minLengthRegex.MatchString(password) {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if !hasNumberRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasUpperRegex.MatchString(password) {
		return fmt.Errorf("password must contain at leat one uppercase letter")
	}
	if !hasLowerRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasSpecialRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
