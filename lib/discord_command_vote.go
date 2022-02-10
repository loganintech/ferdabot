package ferdabot

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
)

var yesnoVoteMenuOption = &discordgo.ApplicationCommand{
	Name:    "Start Yes/No Vote",
	Version: "0.0.1",
	Type:    discordgo.MessageApplicationCommand,
}

var voteCommand = &discordgo.ApplicationCommand{
	Name:        "vote",
	Description: "Start a vote",
	Version:     "0.0.1",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "yesno",
			Description: "Create a yesno vote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "subject",
					Description: "Set the subject of the vote",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "custom",
			Description: "Create a custom vote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "subject",
					Description: "Set the subject of the vote",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "options",
					Description: "Your list of options, separated by emojis.",
					Required:    true,
				},
			},
		},
	},
}

var badVoteStructure = &discordgo.MessageEmbed{
	Type:        discordgo.EmbedTypeRich,
	Title:       "Vote command: bad format",
	Description: "Couldn't parse the vote command arguments. Make sure you're only using an emoji one time.",
	Color:       MODBOT_RED,
}

var reg = regexp.MustCompile(`<?:(?P<name>\w+):(?P<id>\d+)?>?`)

func (b *Bot) handleVoteResponse(_ *discordgo.Session, i *discordgo.Interaction) {
	if i == nil || i.Message == nil || i.Message.ID == "" {
		return
	}
	messageID := i.Message.ID
	messageComponentData := i.MessageComponentData()

	vote, err := b.GetVotePost(messageID)
	if err != nil {
		return
	}

	if !vote.Active {
		if err := b.discord.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: onlyShowSenderFlag,
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Closed Vote",
						Description: "This vote is already closed.",
						Timestamp:   time.Now().Format(time.RFC3339),
						Color:       MODBOT_RED,
					},
				},
			},
		}); err != nil {
			return
		}

		return
	}

	if messageComponentData.CustomID == "close_vote" {
		b.handleCloseVote(vote, i)
	} else {
		if err := b.CastVote(&CastVote{
			MessageID: messageID,
			EmojiName: messageComponentData.CustomID,
			AuthorID:  i.Member.User.ID,
		}); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Bot had error casting vote: %+v\n", err)
			return
		}
	}

	b.postVoteAfterAction(i)
}

func (b *Bot) handleCloseVote(vote *Vote, i *discordgo.Interaction) {
	if vote.CreatorID != i.Member.User.ID {
		if err := b.discord.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: onlyShowSenderFlag,
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Close Error",
						Description: "You may not close this vote because you did not create it.",
						Timestamp:   time.Now().Format(time.RFC3339),
						Color:       MODBOT_RED,
					},
				},
			},
		}); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Bot had error closing vote: %+v\n", err)
			return
		}
		return
	}

	if err := b.CloseVotePost(vote.MessageID); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Bot had error saving vote closed: %+v\n", err)
		return
	}
}

func (b *Bot) postVoteAfterAction(i *discordgo.Interaction) {
	messageID := i.Message.ID

	vote, err := b.GetVotePost(messageID)
	if err != nil {
		return
	}

	lines, err := b.GetLinesForVote(messageID)
	if err != nil {
		return
	}

	castedVotes, err := b.GetCastVotes(messageID)
	if err != nil {
		return
	}

	castVotes := map[string][]*CastVote{}
	for _, castedVote := range castedVotes {
		castVotes[castedVote.EmojiName] = append(castVotes[castedVote.EmojiName], &castedVote)
	}

	embed, rows := b.formatVoteEmbed(vote.Title, vote.Description, lines, castVotes, !vote.Active)

	if err := b.discord.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: rows,
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error responding to discord button press, %w", err)
		return
	}
}

func (b *Bot) handleVoteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var voteTypeString string
	var subject string

	applicationData := i.ApplicationCommandData()
	if applicationData.Name == yesnoVoteMenuOption.Name {
		voteTypeString = "yesno"
	} else {
		voteTypeString = applicationData.Options[0].Name
		subject, _ = applicationData.Options[0].Options[0].Value.(string)
	}

	if applicationData.Name == yesnoVoteMenuOption.Name {
		subject = fmt.Sprintf("[Click to jump to relevant conversation.](https://discord.com/channels/%s/%s/%s)", i.GuildID, i.ChannelID, applicationData.TargetID)
	}

	var lines []Line
	var title string

	switch voteTypeString {
	case "yesno":
		title = fmt.Sprintf("New yes/no vote started by %s", i.Member.User.Username)

		lines = []Line{
			{
				EmojiName: "👍",
				LineValue: "Yes",
			},
			{
				EmojiName: "👎",
				LineValue: "No",
			},
			{
				EmojiName: "🤷",
				LineValue: "Indifferent",
			},
			{
				EmojiName: "🕓",
				LineValue: "Need more time to decide",
			},
		}

	case "custom":
		title = fmt.Sprintf("New custom vote started by %s", i.Member.User.Username)

		options, ok := applicationData.Options[0].Options[1].Value.(string)
		if !ok {
			b.badFormat(s, i)
			return
		}

		var err error
		lines, err = getLines(options)
		if err != nil || len(lines) == 0 {
			b.badFormatErr(s, i.Interaction, "Couldn't parse any options. Make sure to include emojis and descriptions for them.")
			return
		}
	}

	if len(lines) > 24 {
		b.badFormatErr(s, i.Interaction, "You may only include up to 24 options.")
		return
	}

	embed, rows := b.formatVoteEmbed(title, subject, lines, nil, false)

	b.interactionRespond(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: rows,
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})

	response, err := b.discord.InteractionResponse(b.discord.State.User.ID, i.Interaction)
	if err != nil {
		return
	}

	if err := b.CreateNewVote(&Vote{
		MessageID:   response.ID,
		Title:       title,
		Description: subject,
		CreatorID:   i.Member.User.ID,
	}); err != nil {
		return
	}

	for _, line := range lines {
		line.MessageID = response.ID
		if err := b.AddLineToVote(&line); err != nil {
			return
		}
	}
}

func (b *Bot) badFormat(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.interactionRespond(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: onlyShowSenderFlag,
			Embeds: []*discordgo.MessageEmbed{
				badVoteStructure,
			},
		},
	})
}

func (b *Bot) interactionRespond(s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) {
	err := s.InteractionRespond(i, response)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error responding to discord command, %w", err)
	}
}

func (b *Bot) badFormatErr(s *discordgo.Session, i *discordgo.Interaction, description string) {
	b.interactionRespond(s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: onlyShowSenderFlag,
			Embeds: []*discordgo.MessageEmbed{
				{
					Type:        discordgo.EmbedTypeRich,
					Title:       "Create Vote Failed",
					Description: description,
					Timestamp:   time.Now().Format(time.RFC3339),
					Color:       MODBOT_RED,
				},
			},
		},
	})
}
