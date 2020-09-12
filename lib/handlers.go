package ferdabot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// region Event Handlers
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the message if the bot sent it
	if m.Author.ID == s.State.User.ID {
		return
	}

	splitMessage := strings.Split(m.Content, " ")
	command := splitMessage[0]
	data := strings.Join(splitMessage[1:], " ")

	ferdaAction := b.router.ExecuteRoute(command, s, m, data)
	if !ferdaAction.LogOnly {
		if _, err := s.ChannelMessageSend(m.ChannelID, ferdaAction.DiscordText); err != nil {
			fmt.Printf("Error sending message: %s to %s\n", ferdaAction.DiscordText, m.ChannelID)
		}
	}

	fActBytes, _ := json.Marshal(ferdaAction)
	fAct := string(fActBytes)
	fmt.Println(fAct)
}
