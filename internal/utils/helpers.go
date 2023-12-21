package utils

import (
	"net/mail"
	"regexp"
)

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	// Use a regular expression to further validate email format
	if err == nil {
		pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		return regexp.MustCompile(pattern).MatchString(email)
	}

	return false
}

func GetIDFromURL(urlPath string) string {
	// Get the userID from the URL path
	userIDRegex := regexp.MustCompile(`/users/(\d+)`)
	matches := userIDRegex.FindStringSubmatch(urlPath)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}
