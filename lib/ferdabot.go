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

var userRegex = regexp.MustCompile("[0-9]+")
var BannedEchoPhrases = []string{
	"@everyone",
	"@here",
}

type FerdaEntry struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"userid"`
	Reason    string    `db:"reason"`
	When      time.Time `db:"time"`
	CreatorID int64     `db:"creatorid"`
}

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

func (b *Bot) Setup() error {
	// region Discord Conn
	token := os.Getenv("FERDATOKEN")
	if token == "" {
		return fmt.Errorf("A bot environment token is required. FERDATOKEN was empty.\n")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	dg.AddHandler(b.messageCreate)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)
	// endregion

	// region DB
	connURL := os.Getenv("DBCONNURL")
	if connURL == "" {
		return fmt.Errorf("A db connection string is required. DBCONNURL was empty.\n")
	}

	dbCon, sqlErr := sqlx.Connect("postgres", connURL)
	if sqlErr != nil {
		return fmt.Errorf("Error connecting to DB: %s\n", sqlErr)
	}
	// endregion

	// region OS Signal Channel
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	// endregion

	// region Assignments
	b.db = dbCon
	b.dg = dg
	b.signalChannel = sc
	b.treeRouter = NewCommandMatcher()
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
	conErr := b.dg.Open()
	if conErr != nil {
		return fmt.Errorf("Couldn't open discord connection, %s\n", conErr)
	}

	fmt.Println("Startup (ok)")

	<-b.signalChannel
	if closeErr := b.dg.Close(); closeErr != nil {
		return closeErr
	}

	return nil
}

func (b *Bot) ProcessFerdaAction(act FerdaAction, s *discordgo.Session, m *discordgo.MessageCreate) {
	if !act.LogOnly && (s != nil || m != nil) {
		if _, err := s.ChannelMessageSend(m.ChannelID, act.DiscordText); err != nil {
			fmt.Printf("Error sending message: %s to %s\n", act.DiscordText, m.ChannelID)
		}
	}
	if !act.Success() {
		fTreeActBytes, _ := json.Marshal(act)
		fTreeAct := string(fTreeActBytes)
		fmt.Println(fTreeAct)
	}
}
