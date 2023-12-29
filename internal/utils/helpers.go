package utils

import (
	"math/rand"
	"net/mail"
	"regexp"
	"time"
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

func GetIDFromURL(urlPath string) (string, string) {
	var userID string
	var linkID string

	// Get the IDs from the URL path
	userIDRegex := regexp.MustCompile(`/users/(\d+)`)
	userIDMatches := userIDRegex.FindStringSubmatch(urlPath)

	if len(userIDMatches) >= 2 {
		userID = userIDMatches[1]
	}

	linkIDRegex := regexp.MustCompile(`/links/(\d+)`)
	linkIDMatches := linkIDRegex.FindStringSubmatch(urlPath)

	if len(linkIDMatches) >= 2 {
		linkID = linkIDMatches[1]
	}

	return userID, linkID
}

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
