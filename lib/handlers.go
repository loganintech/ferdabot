package ferdabot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// messageCreate handles discordgo.MessageCreate events
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//TODO: Gabe to add code to check if the channel is #aquarium and put a new fish in the aquarium

}

func (b *Bot) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the Message if the bot sent it
	if m.Author != nil && m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "[Original Message Deleted]" {
		if err := s.ChannelMessageDelete(m.ChannelID, m.Message.ID); err != nil {
			b.ProcessFerdaAction(MessageDeleteFailed.RenderLogText(m.ChannelID, m.Message.ID).Finalize(), s, m.Message)
			return
		}
	}
}
