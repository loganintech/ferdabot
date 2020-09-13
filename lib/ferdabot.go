package ferdabot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// Look for a number
var userRegex = regexp.MustCompile("[0-9]+")
var BannedEchoPhrases = []string{
	"@everyone",
	"@here",
}

// FerdaEntry represents the database structure of a ferda entry
type FerdaEntry struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"userid"`
	Reason    string    `db:"reason"`
	When      time.Time `db:"time"`
	CreatorID int64     `db:"creatorid"`
}

// Bot contains or discord and database connections
type Bot struct {
	db            *sqlx.DB
	dg            *discordgo.Session
	signalChannel chan os.Signal
	treeRouter    CommandMatcher
}

func NewBot() Bot {
	return Bot{
		db:            nil,
		dg:            nil,
		signalChannel: nil,
	}
}

// Setup connects to discord and sql
func (b *Bot) Setup() error {
	// region Discord Conn
	// Get the discord token
	token := os.Getenv("FERDATOKEN")
	if token == "" {
		return fmt.Errorf("A bot environment token is required. FERDATOKEN was empty.\n")
	}

	// Set the discord token
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	// Add the messageCreate handler to the bot
	dg.AddHandler(b.messageCreate)
	// And assign the intents
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)
	// endregion

	// region DB
	// Get the DB connection string
	connURL := os.Getenv("DBCONNURL")
	if connURL == "" {
		return fmt.Errorf("A db connection string is required. DBCONNURL was empty.\n")
	}

	// And connect to our postgres instance
	dbCon, sqlErr := sqlx.Connect("postgres", connURL)
	if sqlErr != nil {
		return fmt.Errorf("Error connecting to DB: %s\n", sqlErr)
	}
	// endregion

	// region OS Signal Channel
	// Make a channel for os signals
	sc := make(chan os.Signal, 1)
	// And set the channel to notify on some signals
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	// endregion

	// region Assignments
	b.db = dbCon
	b.dg = dg
	b.signalChannel = sc
	b.treeRouter = NewCommandMatcher()
	// endregion

	// region Routes
	routes := []MessageCreateRoute{
		{key: "!echo", f: b.processEcho, desc: "Echo any message (not in the blacklist)."},
		{key: "+ferda", f: b.processNewFerda, desc: "Add a new ferda with a reason. Ex: `+ferda @Logan for creating ferdabot.`"},
		{key: "?ferda", f: b.processGetFerda, desc: "Get a ferda for a person. Ex: `?ferda @Logan`"},
		{key: "?bigferda", f: b.processDetailedGetFerda, desc: "Get a detailed ferda for a person. Ex: `?bigferda @Logan`"},
		{key: "?ferdasearch", f: b.processSearchFerda, desc: "Search for ferdas for a person containing some text. Ex: `?ferdasearch @Logan ferdabot`"},
		{key: "-ferda", f: b.processRemoveFerda, desc: "Remove a ferda by its ID: `-ferda 7`"},
		{key: "?help", f: b.processHelp, desc: "Sends this help message."},
		{key: "!help", f: b.processHelp, desc: "Sends this help message."},
		{key: "+help", f: b.processHelp, desc: "Sends this help message."},
		{key: "!dice", f: b.processDice, desc: "Roll a dice in the format 1d6."},
	}
	for i, route := range routes {
		if action := b.treeRouter.AddCommand(route.key, &routes[i]); !action.Success() {
			b.ProcessFerdaAction(action, nil, nil)
		}
	}

	// endregion

	return nil
}

func (b *Bot) Run() error {
	// region connecting
	// Start the discord connection and event loop
	conErr := b.dg.Open()
	if conErr != nil {
		return fmt.Errorf("Couldn't open discord connection, %s\n", conErr)
	}

	// Record that the bot has started
	fmt.Println("Startup (ok)")

	// Wait for a signal from the os
	<-b.signalChannel
	// And close the discord connection if we're in shutdown
	if closeErr := b.dg.Close(); closeErr != nil {
		return closeErr
	}

	return nil
}

func (b *Bot) ProcessFerdaAction(act FerdaAction, s *discordgo.Session, m *discordgo.MessageCreate) {
	// If we want to send to discord and have a session / message
	if !act.LogOnly && (s != nil || m != nil) {
		// Send to discord
		if _, err := s.ChannelMessageSend(m.ChannelID, act.DiscordText); err != nil {
			fmt.Printf("Error sending message: %s to %s\n", act.DiscordText, m.ChannelID)
		}
	}

	// If the message was not a success
	if !act.Success() {
		// Marshal the action to string and print
		fTreeActBytes, _ := json.Marshal(act)
		fTreeAct := string(fTreeActBytes)
		fmt.Println(fTreeAct)
	}
}
