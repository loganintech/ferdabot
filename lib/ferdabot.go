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

// ConfigEntry is a struct to represent configuration parameters
type ConfigEntry struct {
	Key string `db:"key"`
	Val string `db:"val"`
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
	discord       *discordgo.Session
	signalChannel chan os.Signal
}

func NewBot() Bot {
	return Bot{
		db:            nil,
		discord:       nil,
		signalChannel: nil,
	}
}

// Setup connects to discord and sql
func (b *Bot) Setup() (func(), error) {
	if err := b.setupDBConnection(); err != nil {
		return nil, err
	}

	cleanupDiscord, err := b.setupDiscord()
	if err != nil {
		return nil, err
	}
	b.setupSignalHandler()
	go b.reminderLoop()

	return func() {
		cleanupDiscord()
	}, nil
}

func (b *Bot) setupDBConnection() error {
	connURL := os.Getenv("DBCONNURL")
	if connURL == "" {
		return fmt.Errorf("A db connection string is required. DBCONNURL was empty.\n")
	}
	if err := b.setupPostgres(connURL); err != nil {
		return err
	}
	return nil
}

func (b *Bot) setupPostgres(connURL string) error {
	dbCon, sqlErr := sqlx.Connect("postgres", connURL)
	if sqlErr != nil {
		return fmt.Errorf("Error connecting to DB: %s\n", sqlErr)
	}
	b.db = dbCon
	return nil
}

func (b *Bot) setupSignalHandler() {
	// Make a channel for os signals
	sc := make(chan os.Signal, 1)
	// And set the channel to notify on some signals
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	b.signalChannel = sc
}

func (b *Bot) Run() error {
	// Wait for a signal from the os
	<-b.signalChannel
	// And close the discord connection if we're in shutdown
	if closeErr := b.discord.Close(); closeErr != nil {
		return closeErr
	}

	return nil
}

func (b *Bot) ProcessFerdaAction(act Action, s *discordgo.Session, m *discordgo.Message) {
	if act.DontLog {
		return
	}

	// If we want to send to discord and have a session / Message
	if !act.LogOnly && (s != nil || m != nil) {
		// Send to discord
		if _, err := s.ChannelMessageSend(m.ChannelID, act.DiscordText); err != nil {
			fmt.Printf("Error sending Message: %s to %s\n", act.DiscordText, m.ChannelID)
		}
	}

	// If the Message was not a success
	if !act.Success() {
		// Marshal the action to string and print
		fTreeActBytes, _ := json.Marshal(act)
		fTreeAct := string(fTreeActBytes)
		fmt.Println(fTreeAct)
	}
}
