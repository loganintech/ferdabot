package ferdabot

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) setupDiscord() (func(), error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("A bot environment token is required. FERDATOKEN was empty.\n")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	dg.AddHandler(b.messageCreate)
	dg.AddHandler(b.messageUpdate)

	// And assign the intents
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)
	b.discord = dg

	applicationCommandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction{
		"dice":   b.handleRollDiceCommand,
		"choice": b.handleChoiceCommand,
		"ping":   b.handlePingCommand,
		"remind": b.handleRemindCommand,
		"ferda":  b.handleFerdaCommand,
	}

	b.discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := applicationCommandHandlers[i.ApplicationCommandData().Name]; ok {
			respondWith := "Unknown error, please contact Logan (JewsOfHazard)"
			action := handler(s, i)
			respondWith = action.DiscordText

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: respondWith,
				},
			})
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "error responding to discord command, %w", err)
			}

			if !action.DontLog {
				fmt.Println(action.LogText)
			}
		}
	})

	b.discord.AddHandler(func(s *discordgo.Session, i *discordgo.Ready) {
		fmt.Println("Discord is up!")
	})

	conErr := b.discord.Open()
	if conErr != nil {
		return nil, fmt.Errorf("Couldn't open discord connection, %s\n", conErr)
	}

	if _, err := b.discord.ApplicationCommandBulkOverwrite(b.discord.State.Ready.User.ID, os.Getenv("DISCORD_GUILD_ID"), []*discordgo.ApplicationCommand{
		diceCommand,
		//choiceCommand,
		pingCommand,
		//ferdaCommand,
		//remindCommand,
	}); err != nil {
		return nil, err
	}

	return func() {
		created, err := b.discord.ApplicationCommands(b.discord.State.Ready.User.ID, os.Getenv("DISCORD_GUILD_ID"))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "couldn't get discord commands, %w", err)
		}
		for _, route := range created {
			if err := b.discord.ApplicationCommandDelete(b.discord.State.Ready.User.ID, os.Getenv("DISCORD_GUILD_ID"), route.ID); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "couldn't delete discord commands, %w", err)
			}
		}

		_ = b.discord.Close()
	}, nil
}

func (b *Bot) handleRollDiceCommand(_ *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction {
	return b.processDice(i.ApplicationCommandData().Options[0].Value.(string))
}

var diceCommand = &discordgo.ApplicationCommand{
	Name:        "roll",
	Description: "Roll some dice.",
	Version:     "0.0.1",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "dice",
			Description: "Dice roll format. Ex: 1d6, 3d6, 2d3+2",
			Required:    true,
		},
	},
}

func (b *Bot) handleChoiceCommand(_ *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction {
	return DontLog
}

var choiceCommand = &discordgo.ApplicationCommand{}

func (b *Bot) handlePingCommand(_ *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction {
	return b.processPing(i.Message)
}

var pingCommand = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Pong! (Shows Bot Network Latency)",
	Version:     "0.0.1",
	Type:        discordgo.ChatApplicationCommand,
}

func (b *Bot) handleFerdaCommand(_ *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction {
	//	{key: "+ferda", f: b.processNewFerda, desc: "Add a new ferda with a reason. Ex: `+ferda @Logan for creating ferdabot.`"},
	//	{key: "?ferda", f: b.processGetFerda, desc: "Get a ferda for a person. Ex: `?ferda @Logan`"},
	//	{key: "-ferda", f: b.processDeleteFerda, desc: "Remove a ferda by its ID: `-ferda 7`"},
	//	{key: "?bigferda", f: b.processDetailedGetFerda, desc: "Get a detailed ferda for a person. Ex: `?bigferda @Logan`"},
	//	{key: "?ferdasearch", f: b.processSearchFerda, desc: "Search for ferdas for a person containing some text. Ex: `?ferdasearch @Logan ferdabot`"},
	return DontLog
}

var ferdaCommand = &discordgo.ApplicationCommand{}

func (b *Bot) handleRemindCommand(_ *discordgo.Session, i *discordgo.InteractionCreate) FerdaAction {
	//	{key: "+remindme", f: b.processNewReminder, desc: "Creates a new reminder at a certain time."},
	//	{key: "?remindme", f: b.processGetReminders, desc: "DMs you a list of your reminders."},
	//	{key: "-remindme", f: b.processDeleteReminder, desc: "Deletes a reminder by ID."},

	return DontLog
}

var remindCommand = &discordgo.ApplicationCommand{}
