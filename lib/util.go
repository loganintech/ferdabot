package ferdabot

import "strings"

// getUserFromFirstWord returns a username string from the first arg in text
func (b *Bot) getUserFromFirstWord(trimmedText string) string {
	// Split text into words
	split := strings.Split(trimmedText, " ")
	user := split[0]

	// Use the userRegex to find the number from the username
	found := userRegex.Find([]byte(user))
	foundString := string(found)
	// And return it
	return foundString
}
