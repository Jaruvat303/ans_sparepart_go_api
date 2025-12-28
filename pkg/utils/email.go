package utils

import (
	"regexp"
	"strings"
)

// Define a regex for email validation
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func IsValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
} // ตรวจสอบ email

// MaskEmail masks the user part of an email address.
// It keeps the first and last character of the username and the entire domain.
// Example: "john.doe@example.com" -> "j***e@example.com"
func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}

	// หาตำแหน่งของ @ หากไม่พบจะ return -1
	atIndax := strings.LastIndex(email, "@")
	if atIndax == -1 {
		return "Invalid email format: missing '@'"
	}

	// แยก username กับ domain email
	username := email[:atIndax]
	domain := email[atIndax+1:]

	if len(username) <= 2 {
		return username + "@" + domain // หาก email สั้นไปไม่ต้อง mark
	}

	// สร้าง mark username โดย อักษรแรกของ username + สร้าง * ตามจำนวนชื่อผู้ใช้ -2 (ตัวแรกและตัวสุดท้าย) + ตัวอักษรสุดทัาย username
	markedUsername := username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]

	// ประกอบ maekedUsername กับ Doamin Email
	return markedUsername + "@" + domain
}
