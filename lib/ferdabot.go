package ferdabot

import (
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
	router        MessageCreateRouter
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
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
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
	// endregion

	if setupErr := b.SetupRoutes(); setupErr != nil {
		return setupErr
	}

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

func (b *Bot) SetupRoutes() error {
	b.router = NewMessageCreateRouter()

	if ferdaAction := b.router.AddRoute("!echo", b.processEcho); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRoute("+ferda", b.processNewFerda); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRoute("?ferda", b.processGetFerda); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRouteWithAliases([]string{"?help", "!help", "+help"}, processHelp); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRoute("?bigferda", b.processDetailedGetFerda); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRoute("?ferdasearch", b.processSearchFerda); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}
	if ferdaAction := b.router.AddRoute("-ferda", b.processRemoveFerda); !ferdaAction.Success() {
		return fmt.Errorf(ferdaAction.LogText)
	}

	return nil
}
