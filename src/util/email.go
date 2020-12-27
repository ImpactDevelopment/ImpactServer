package util

import "regexp"

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+\\/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// IsValidEmail checks if the email provided passes the required structure and length.
func IsValidEmail(email string) bool {
	if len(email) < 3 && len(email) > 254 {
		return false
	}
	return emailPattern.MatchString(email)
}
