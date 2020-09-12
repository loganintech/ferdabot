package ferdabot

// FerdaMessages
var (
	// Success
	HelpMessage     = FerdaSuccess("Use `+ferda @User [reason]` to add a ferda, and `?ferda @User` to get a ferda.", "Help message sent.")
	NewFerda        = FerdaSuccess("Thanks for the new ferda <@!%s>", "Thanked %s for the ferda.")
	NotFerdaMessage = FerdaSuccess("<@!%s> is not ferda.", "%s - %s is not ferda.")
	GetFerda        = FerdaSuccess("<@!%s> was ferda on %s for %s.", "Error sending message to discord %s.")
	EchoSuccess     = FerdaSuccess("%s", "Echo'd ferda message")

	// Failure
	PingFailure     = FerdaFailure("You must ping someone who is ferda. Ex: `+ferda @Logan is a great guy`", "Failed to find a username in message %s.")
	DBInsertErr     = FerdaFailure("Database error occurred, please contact Logan.", "Error inserting into the DB %+v")
	NoRowDBErr      = FerdaFailure("Database error occurred, please contact Logan.", "No rows affected by DB insert.")
	UserNotFoundErr = FerdaFailure("That user was not found.", "Couldn't load user from discord: %s")
	DBGetErr        = FerdaFailure("Database error occurred, please contact Logan.", "Error getting from table %+v")
	EchoFailure     = FerdaFailure("You shouldn't be trying to send that you dumb fuck.", "User: %s, ID: %s tried to send %s")
)
