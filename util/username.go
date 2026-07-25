package util

import "regexp"

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)

func IsValidUsername(s string) bool {
	if len(s) < 8 || len(s) > 16 {
		return false
	}
	return usernamePattern.MatchString(s)
}
