package ferdabot

import (
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

	b.processCommands(s, m)
}
