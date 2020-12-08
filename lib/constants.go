package ferdabot

// FerdaMessages
var (
	// Success
	NewFerda         = FerdaSuccess("Thanks for the new ferda <@!%s>", "Thanked %s for the ferda.")
	NotFerdaMessage  = FerdaSuccess("<@!%s> is not ferda.", "%s - %s is not ferda.")
	GetFerda         = FerdaSuccess("<@!%s> was ferda on %s for %s.", "Sent a retrieved ferda message %s.")
	EchoSuccess      = FerdaSuccess("%s", "Echo'd ferda message")
	GetDetailedFerda = FerdaSuccess("[%d - %s] <@!%s> was ferda on %s for %s added by <@!%d>.", "Sent a detailed ferda message %s.")
	DeletedFerda     = FerdaSuccess("Deleted a ferda successfully <@!%s>. ID: [%s]", "Deleted a ferda with id %d.")
	HelpBody         = FerdaSuccess("> [%s] %s", "Help message body.")
	DiceBody         = FerdaSuccess("Rolling [%dd%d%s]: %d", "Rolling [%dd%d%s]: %d")
	ChoiceResult     = FerdaSuccess("The choice is: %s", "Chose %s from random array %v")

	// Failure
	MentionMissing     = FerdaFailure("You must ping an associated user. Ex: `+ferda @Logan is a great guy` or `?ferda @Logan`", "Failed to find a username in message %s.")
	DBInsertErr        = FerdaFailure("Database error occurred, please contact Logan.", "Error inserting into the DB %+v")
	DBDeleteErr        = FerdaFailure("Database error occurred, please contact Logan.", "Error deleting frmo the DB %+v.")
	UserNotFoundErr    = FerdaFailure("That user was not found.", "Couldn't load user from discord: %s")
	DBGetErr           = FerdaFailure("Database error occurred, please contact Logan.", "Error getting from table %+v")
	EchoFailure        = FerdaFailure("You shouldn't be trying to send that you dumb fuck.", "User: %s, ID: %s tried to send %s")
	MissingID          = FerdaFailure("You must include a ferda id to delete `-ferda [id]`", "Not given an ID of ferda to delete.")
	NotEnoughArguments = FerdaFailure("Not enough arguments, need %d minimum.", "Not enough arguments supplied to command %s.")
	BadDiceFormat      = FerdaFailure("Bad dice format. Please use the form NdN (1d6 for example).", "Bad dice format %s")
	BadIDFormat        = FerdaFailure("ID was not in the correct format: %s", "ID was not in the correct format: %s")

	// LogOnly Fail
	AddRouteFailed             = FerdaLogOnly("Adding route failed, %s already exists").SetFail()
	RouteNotFound              = FerdaLogOnly("Route %s was not found").SetFail()
	SpotifyPlaylistLoadError   = FerdaLogOnly("Spotify Playlist not Found [%v]").SetFail()
	SpotifyCreatePlaylistError = FerdaLogOnly("Spotify Playlist Creation Failed [%v]").SetFail()
	SpotifyAddToPlaylistError  = FerdaLogOnly("Spotify Tracks failed to Add to Playlist [%v]").SetFail()
	SpotifyUserNotFound        = FerdaLogOnly("Current Spotify user Not Set [%v]").SetFail()
	ParamNotFound              = FerdaLogOnly("ConfigEntry was not found: %s").SetFail()

	// LogOnly Success
	SpotifySongAdded = FerdaLogOnly("Song [%s] Added to Playlist: %s").SetSuccess()
	AddRouteSuccess  = FerdaLogOnly("Adding route complete: %s").SetSuccess()

	// Pre-Finalized Success
	HelpHeader          = FerdaSuccess("Here's the list of available commands:", "Help message sent.").Finalize()
	DBSuccess           = FerdaLogOnly("DB Success").SetSuccess().Finalize()
	MultipleFerdasFound = FerdaSuccess("Found the following ferdas:", "Found multiple ferdas:").SetSuccess().Finalize()
	DiceHeader          = FerdaSuccess("Rolling dice:", "Rolling Dice:").Finalize()
	//RoutesFinished      = FerdaLogOnly("Routes finished setup without error.").SetSuccess().Finalize()

	// Pre-Finalized Failure
	NoRowDBErr              = FerdaFailure("Database error occurred, please contact Logan.", "No rows affected in DB.").Finalize()
	DBResultsEmpty          = FerdaFailure("No entries found.", "No rows returned in result set.")
	TooManyDiceToRoll       = FerdaFailure("You can't roll that many dice.", "User wanted to roll too many dice.").SetFail().Finalize()
	SpotifyPlaylistNotFound = FerdaLogOnly("Spotify Playlist not Found").SetFail().Finalize()
	FromChannelNotFound     = FerdaLogOnly("From Channel not Found").SetFail().Finalize()
	PlaylistIDChannelClosed = FerdaLogOnly("PlaylistID Channel Not Closed").SetFail().Finalize()

	DontLog = FerdaActionBuilder{DontLog: true}.Finalize()
)
