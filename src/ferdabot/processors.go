package ferdabot

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) processCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message %s\n", m.Content)
	splitMessage := strings.Split(m.Content, " ")
	command := splitMessage[0]
	data := strings.Join(splitMessage[1:], " ")
	switch command {
	case "!echo":
		b.processEcho(s, m, data)
	case "+ferda":
		b.processNewFerda(s, m, data)
	case "?ferda":
		b.processGetFerda(s, m, data)
	case "?help", "!help", "+help":
		if _, err := s.ChannelMessageSend(m.ChannelID, "Use `+ferda @User [reason]` to add a ferda, and `?ferda @User` to get a ferda."); err != nil {
			fmt.Printf("Error sending message to discord %s\n", err)
		}
	}
}

func (b *Bot) processEcho(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	for _, phrase := range BannedEchoPhrases {
		if strings.Contains(trimmedText, phrase) {
			if _, err := s.ChannelMessageSend(m.ChannelID, "You shouldn't be trying to send that you dumb fuck."); err != nil {
				fmt.Printf("Error occured responding to ping. %s\n", err)
			}
			return
		}
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, trimmedText); err != nil {
		fmt.Printf("Error occured responding to ping. %s\n", err)
	}
}

func (b *Bot) processNewFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	split := strings.Split(trimmedText, " ")
	foundString := b.getUserFromText(trimmedText)

	if foundString == "" {
		if _, err := s.ChannelMessageSend(m.ChannelID, "You must ping someone who is ferda. Ex: `+ferda @Logan is a great guy`"); err != nil {
			fmt.Printf("Error occured responding to ping. %s\n", err)
		}
		return
	}

	res, dbErr := b.db.NamedExec(`INSERT INTO ferda (userid, time, reason, creatorid) VALUES (:userid, :time, :reason, :creatorid)`, map[string]interface{}{
		"userid": foundString,
		// Adjust for local time of bot
		"time":      time.Now().Round(time.Microsecond).Add(-(time.Hour * 7)),
		"reason":    strings.Join(split[1:], " "),
		"creatorid": m.Author.ID,
	})
	if dbErr != nil {
		fmt.Printf("Error inserting into the DB %s\n", dbErr)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		if _, err := s.ChannelMessageSend(m.ChannelID, "No rows effected."); err != nil {
			fmt.Printf("Error sending a message to discord.")
			return
		}
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thanks for the new ferda <@!%s>", m.Author.ID)); err != nil {
		fmt.Printf("Error sending a message to discord.")
		return
	}
}

func (b *Bot) processGetFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	foundUser := b.getUserFromText(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		fmt.Printf("Couldn't load user from discord: %s\n", err)
	}

	ferdaEntry := FerdaEntry{}
	dbErr := b.db.Get(&ferdaEntry, `SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`, foundUser)
	if dbErr != nil {
		if dbErr.Error() == "sql: no rows in result set" {
			if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@!%s> is not ferda.", user.ID)); err != nil {
				fmt.Printf("Error sending message to discord %s\n", err)
			}
		}
		fmt.Printf("Error selecting from table %+v\n", dbErr)
		return
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@!%s> was ferda on %s for %s", user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006"), ferdaEntry.Reason)); err != nil {
		fmt.Printf("Error sending message to discord %s\n", err)
	}
}
