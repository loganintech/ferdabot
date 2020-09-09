package main

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
	_ "github.com/lib/pq"
)

var db *sqlx.DB
var userRegex = regexp.MustCompile("[0-9]+")

type FerdaEntry struct {
	ID     int64     `db:"id"`
	UserID int64     `db:"userid"`
	Reason string    `db:"reason"`
	When   time.Time `db:"time"`
}

func main() {
	token := os.Getenv("FERDATOKEN")
	if token == "" {
		fmt.Println("A bot environment token is required. FERDATOKEN was empty.")
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Printf("Error occurred starting the bot %s\n", err)
		os.Exit(1)
	}

	// region config
	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	// endregion

	dbCon, sqlErr := sqlx.Connect("postgres", "host=db user=postgres dbname=postgres password=pass sslmode=disable")
	if sqlErr != nil {
		fmt.Printf("Error connecting to DB: %s\n", sqlErr)
	}

	db = dbCon

	// region connecting
	conErr := dg.Open()
	if conErr != nil {
		fmt.Println("Couldn't open discord connection.")
		os.Exit(2)
	}

	fmt.Println("Startup (ok)")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	<-sc
	if closeErr := dg.Close(); closeErr != nil {
		fmt.Printf("Error closing connection %s\n", closeErr.Error())
		os.Exit(3)
	}
	//endregion connecting
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	m.Content = strings.TrimSpace(m.Content)
	// Ignore the message if the bot sent it or if it's not a command
	if m.Author.ID == s.State.User.ID {
		return
	}

	handleCommands(s, m)

}

func handleCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message %s\n", m.Content)
	splitMessage := strings.Split(m.Content, " ")
	command := splitMessage[0]
	data := strings.Join(splitMessage[1:], " ")
	switch command {
	case "!echo":
		handleEcho(s, m, data)
	case "!ferda":
		handleNewFerda(s, m, data)
	case "?ferda":
		handleGetFerda(s, m, data)
	case "?help", "!help":
		if _, err := s.ChannelMessageSend(m.ChannelID, "Use `!ferda @User [reason]` to add a ferda, and `?ferda @User` to get a ferda."); err != nil {
			fmt.Printf("Error sending message to discord %s\n", err)
		}
	}
}

var BANNED_ECHO_PHRASES []string = []string{
	"@everyone",
	"@here",
}

func handleEcho(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	for _, phrase := range BANNED_ECHO_PHRASES {
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

func getUserFromText(trimmedText string) string {
	split := strings.Split(trimmedText, " ")
	user := split[0]

	found := userRegex.Find([]byte(user))
	foundString := string(found)
	return foundString
}

func handleNewFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	split := strings.Split(trimmedText, " ")
	foundString := getUserFromText(trimmedText)

	if foundString == "" {
		if _, err := s.ChannelMessageSend(m.ChannelID, "You must ping someone who is ferda. Ex: `!ferda @Logan is a great guy`"); err != nil {
			fmt.Printf("Error occured responding to ping. %s\n", err)
		}
		return
	}

	res, dbErr := db.NamedExec(`INSERT INTO ferda (userid, time, reason) VALUES (:userid, :time, :reason)`, map[string]interface{}{
		"userid": foundString,
		"time":   time.Now(),
		"reason": strings.Join(split[1:], " "),
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
}

func handleGetFerda(s *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) {
	foundUser := getUserFromText(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		fmt.Printf("Couldn't load user from discord: %s\n", err)
	}

	ferdaEntry := FerdaEntry{}
	dbErr := db.Get(&ferdaEntry, `SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`, foundUser)
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
