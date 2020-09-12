package ferdabot

import "strings"

func (b *Bot) getUserFromFirstWord(trimmedText string) string {
	split := strings.Split(trimmedText, " ")
	user := split[0]

	found := userRegex.Find([]byte(user))
	foundString := string(found)
	return foundString
}
