package ferdabot

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
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
}

func NewBot() Bot {
	return Bot{
		db:            nil,
		dg:            nil,
		signalChannel: nil,
	}
}

func (b *Bot) Setup() error {
	token := os.Getenv("FERDATOKEN")
	if token == "" {
		fmt.Println("A bot environment token is required. FERDATOKEN was empty.")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	// region config
	dg.AddHandler(b.messageCreate)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	// endregion

	dbCon, sqlErr := sqlx.Connect("postgres", "host=db user=postgres dbname=postgres password=pass sslmode=disable")
	if sqlErr != nil {
		fmt.Printf("Error connecting to DB: %s\n", sqlErr)
	}

	b.db = dbCon
	b.dg = dg

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	b.signalChannel = sc

	return nil
}

func (b *Bot) Run() error {
	// region connecting
	conErr := b.dg.Open()
	if conErr != nil {
		fmt.Println("Couldn't open discord connection.")
		os.Exit(2)
	}

	fmt.Println("Startup (ok)")

	<-b.signalChannel
	if closeErr := b.dg.Close(); closeErr != nil {
		return closeErr
	}

	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the message if the bot sent it
	if m.Author.ID == s.State.User.ID {
		return
	}

	b.handleCommands(s, m)

}

func (b *Bot) handleCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message %s\n", m.Content)
	splitMessage := strings.Split(m.Content, " ")
	command := splitMessage[0]
	data := strings.Join(splitMessage[1:], " ")
	switch command {
	case "!echo":
		b.handleEcho(s, m, data)
	case "+ferda":
		b.handleNewFerda(s, m, data)
	case "?ferda":
		b.handleGetFerda(s, m, data)
	case "?help", "!help", "+help":
		if _, err := s.ChannelMessageSend(m.ChannelID, "Use `+ferda @User [reason]` to add a ferda, and `?ferda @User` to get a ferda."); err != nil {
			fmt.Printf("Error sending message to discord %s\n", err)
		}
	}
}

func (b *Bot) handleEcho(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	for _, phrase := range BannedEchoPhrases {
		if strings.Contains(trimmedText, phrase) {
			if _, err := s.ChannelMessageSend(m.ChannelID, "You shouldn't be trying to send that you dumb fuck."); err != nil {
				fmt.Printf("Error occured responding to ping. %s\n", err)
			}
			return
		}
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, trimmedText); err != nil {
		fmt.Printf("Error occured responding to ping. %s\n", err)
	}
}

func (b *Bot) getUserFromText(trimmedText string) string {
	split := strings.Split(trimmedText, " ")
	user := split[0]

	found := userRegex.Find([]byte(user))
	foundString := string(found)
	return foundString
}

func (b *Bot) handleNewFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	split := strings.Split(trimmedText, " ")
	foundString := b.getUserFromText(trimmedText)

	if foundString == "" {
		if _, err := s.ChannelMessageSend(m.ChannelID, "You must ping someone who is ferda. Ex: `!ferda @Logan is a great guy`"); err != nil {
			fmt.Printf("Error occured responding to ping. %s\n", err)
		}
		return
	}

	res, dbErr := b.db.NamedExec(`INSERT INTO ferda (userid, time, reason, creatorid) VALUES (:userid, :time, :reason, :creatorid)`, map[string]interface{}{
		"userid":    foundString,
		"time":      time.Now(),
		"reason":    strings.Join(split[1:], " "),
		"creatorid": m.Author.ID,
	})
	if dbErr != nil {
		fmt.Printf("Error inserting into the DB %s\n", dbErr)
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		if _, err := s.ChannelMessageSend(m.ChannelID, "No rows effected."); err != nil {
			fmt.Printf("Error sending a message to discord.")
			return
		}
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thanks for the new ferda <@!%s>", m.Author.ID)); err != nil {
		fmt.Printf("Error sending a message to discord.")
		return
	}
}

func (b *Bot) handleGetFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	foundUser := b.getUserFromText(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		fmt.Printf("Couldn't load user from discord: %s\n", err)
	}

	ferdaEntry := FerdaEntry{}
	dbErr := b.db.Get(&ferdaEntry, `SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`, foundUser)
	if dbErr != nil {
		if dbErr.Error() == "sql: no rows in result set" {
			if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@!%s> is not ferda.", user.ID)); err != nil {
				fmt.Printf("Error sending message to discord %s\n", err)
			}
		}
		fmt.Printf("Error selecting from table %+v\n", dbErr)
		return
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@!%s> was ferda on %s for %s", user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006"), ferdaEntry.Reason)); err != nil {
		fmt.Printf("Error sending message to discord %s\n", err)
	}
}
