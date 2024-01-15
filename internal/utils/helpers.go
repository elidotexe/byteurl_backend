package utils

import (
	"encoding/base64"
	"math/rand"
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

func GenerateRandomHash(length int) (string, error) {
	byteLength := (length * 6) / 8
	if (length*6)%8 != 0 {
		byteLength++
	}

	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	randomString = randomString[:length]

	return randomString, nil
}
