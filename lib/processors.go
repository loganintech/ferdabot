package ferdabot

import (
	"encoding/json"
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

	var ferdaAction FerdaAction
	switch command {
	case "!echo":
		ferdaAction = b.processEcho(m, data)
	case "+ferda":
		ferdaAction = b.processNewFerda(m, data)
	case "?ferda":
		ferdaAction = b.processGetFerda(s, data)
	case "?help", "!help", "+help":
		ferdaAction = HelpMessage
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, ferdaAction.DiscordText); err != nil {
		fmt.Printf("Error sending message: %s to %s\n", ferdaAction.DiscordText, m.ChannelID)
	}

	fmt.Println(json.Marshal(ferdaAction))
}

func (b *Bot) processEcho(m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	for _, phrase := range BannedEchoPhrases {
		if strings.Contains(trimmedText, phrase) {
			return EchoFailure.RenderLogText(m.Author.Username, m.Author.ID, trimmedText)
		}
	}

	return EchoSuccess.RenderDiscordText(trimmedText)
}

func (b *Bot) processNewFerda(m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	split := strings.Split(trimmedText, " ")
	foundString := b.getUserFromText(trimmedText)

	if foundString == "" {
		return PingFailure.RenderLogText(m.Author.Username)
	}

	res, dbErr := b.db.NamedExec(`INSERT INTO ferda (userid, time, reason, creatorid) VALUES (:userid, :time, :reason, :creatorid)`, map[string]interface{}{
		"userid": foundString,
		// Adjust for local time of bot
		"time":      time.Now().Round(time.Microsecond).Add(-(time.Hour * 7)),
		"reason":    strings.Join(split[1:], " "),
		"creatorid": m.Author.ID,
	})
	if dbErr != nil {
		return DBInsertErr.RenderLogText(dbErr)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return NoRowDBErr
	}

	return NewFerda.RenderDiscordText(m.Author.ID).RenderLogText(m.Author.Username)
}

func (b *Bot) processGetFerda(s *discordgo.Session, trimmedText string) FerdaAction {
	foundUser := b.getUserFromText(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err)
	}

	ferdaEntry := FerdaEntry{}
	dbErr := b.db.Get(&ferdaEntry, `SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`, foundUser)
	if dbErr != nil {
		if dbErr.Error() == "sql: no rows in result set" {
			return NotFerdaMessage.RenderDiscordText(user.ID).RenderLogText(user.Username, user.ID)
		}
		return DBGetErr.RenderLogText(dbErr)
	}

	return GetFerda.RenderDiscordText(user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006"), ferdaEntry.Reason).RenderLogText(err)
}
