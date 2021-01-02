package ferdabot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// region Event Handlers
// messageCreate handles discordgo.MessageCreate events
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the Message if the bot sent it
	if m.Author.ID == s.State.User.ID {
		return
	}

	if spotifyAction := b.processSpotifyLink(s, m); spotifyAction != nil {
		b.ProcessFerdaAction(*spotifyAction, s, m.Message)
		return
	}

	// Split our command out
	splitMessage := strings.Split(m.Content, " ")
	command := splitMessage[0]
	data := strings.Join(splitMessage[1:], " ")

	// And execute the command found
	treeFerdaAction := b.treeRouter.ExecuteCommand(command, s, m, data)
	// Then process the result
	b.ProcessFerdaAction(treeFerdaAction, s, m.Message)
}

func (b *Bot) messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the Message if the bot sent it
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "[Original Message Deleted]" {
		if err := s.ChannelMessageDelete(m.ChannelID, m.Message.ID); err != nil {
			b.ProcessFerdaAction(MessageDeleteFailed.RenderLogText(m.ChannelID, m.Message.ID).Finalize(), s, m.Message)
			return
		}
	}
}

// endregion
