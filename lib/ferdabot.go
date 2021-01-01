package ferdabot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/zmb3/spotify"
)

const (
	redirectURI       = "http://localhost:8080/callback"
	spotifyEnabledKey = "spotifyEnabled"
	spotifyCodeKey    = "spotifyCode"
)

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadCollaborative)
	ch    = make(chan *spotify.Client)
	state = "awd123"
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
	db                *sqlx.DB
	dg                *discordgo.Session
	signalChannel     chan os.Signal
	treeRouter        CommandMatcher
	spotifyConnection *spotify.Client
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
	b.dg = dg
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
	b.db = dbCon
	// endregion

	// region OS Signal Channel
	// Make a channel for os signals
	sc := make(chan os.Signal, 1)
	// And set the channel to notify on some signals
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	// endregion

	//region Spotify
	b.initializeSpotify()
	//endregion

	// region Assignments
	b.signalChannel = sc
	b.treeRouter = NewCommandMatcher()
	// endregion

	// region Routes
	routes := []MessageCreateRoute{
		{key: "+ferda", f: b.processNewFerda, desc: "Add a new ferda with a reason. Ex: `+ferda @Logan for creating ferdabot.`"},
		{key: "?ferda", f: b.processGetFerda, desc: "Get a ferda for a person. Ex: `?ferda @Logan`"},
		{key: "-ferda", f: b.processDeleteFerda, desc: "Remove a ferda by its ID: `-ferda 7`"},
		{key: "?bigferda", f: b.processDetailedGetFerda, desc: "Get a detailed ferda for a person. Ex: `?bigferda @Logan`"},
		{key: "?ferdasearch", f: b.processSearchFerda, desc: "Search for ferdas for a person containing some text. Ex: `?ferdasearch @Logan ferdabot`"},
		{key: "!choice", f: b.processChoice, desc: "Choose a random item from a list. Format `!choose Item1|Item2 | Item3| Item 4 | ...`"},
		{key: "!dice", f: b.processDice, desc: "Roll a dice in the format 1d6."},
		{key: "!echo", f: b.processEcho, desc: "Echo any Message (not in the blacklist)."},
		{key: "+help", f: b.processHelp, desc: "Sends this help Message."},
		{key: "?help", f: b.processHelp, desc: "Sends this help Message."},
		{key: "!help", f: b.processHelp, desc: "Sends this help Message."},
		{key: "+remindme", f: b.processNewReminder, desc: "Creates a new reminder at a certain time."},
		{key: "?remindme", f: b.processGetReminders, desc: "DMs you a list of your reminders."},
		{key: "-remindme", f: b.processDeleteReminder, desc: "Deletes a reminder by ID."},
	}
	for i, route := range routes {
		if action := b.treeRouter.AddCommand(route.key, &routes[i]); !action.Success() {
			b.ProcessFerdaAction(action, nil, nil)
		}
	}
	// endregion

	//region Reminders
	go b.reminderLoop()
	//endregion

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

func (b *Bot) initializeSpotify() {
	spotifyEnabledStr, dbAction := b.getConfigEntry(spotifyEnabledKey)
	if !dbAction.Success() {
		b.ProcessFerdaAction(dbAction, nil, nil)
	}
	if dbAction.DBNotFound() {
		b.insertConfigEntry(spotifyEnabledKey, strconv.FormatBool(true))
	}
	spotifyEnabled, configErr := strconv.ParseBool(spotifyEnabledStr.Val)
	spotifyEnabled = spotifyEnabled && configErr == nil

	if spotifyEnabled {
		spotifyOAuthCode, codeAction := b.getConfigEntry(spotifyCodeKey)
		if codeAction.Success() {
			spotifyToken, tokenErr := auth.Exchange(spotifyOAuthCode.Val)
			if tokenErr == nil {
				spotifyClient := auth.NewClient(spotifyToken)
				b.spotifyConnection = &spotifyClient
			}
		}
		if b.spotifyConnection == nil {
			http.HandleFunc("/callback", b.performFirstTimeSpotifyAuth)
			go http.ListenAndServe(":8080", nil)
			url := auth.AuthURL(state)
			fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
			spotifyClient := <-ch
			b.spotifyConnection = spotifyClient
		}
	}
}

func (b *Bot) performFirstTimeSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if code := r.FormValue("code"); code != "" {
		insertAct := b.insertConfigEntry(spotifyCodeKey, code)
		b.ProcessFerdaAction(insertAct, nil, nil)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	w.Header().Set("Content-Type", "text/html")
	ch <- &client
}
