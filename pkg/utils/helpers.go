package utils

import (
	"strings"
	"time"
)

// Default pagination values
const (
	defaultLimit = 10
	maxLimit     = 100
)

// Time Utilities:
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
} // ปัจจุบัน timestamp ISO format

// String Utilities:

// ลบ HTML tags และ XSS
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)
	// Replace multiple spaces with a single space
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// Mark Token
func MarkToken(token string) string {
	if len(token) <= 8 {
		return "*****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// Pagination Utilities:
func NormalizePagination(limit, offset int) (int, int) {
	if limit < 1 {
		limit = defaultLimit

	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
} // ปรับ pagination parameters
