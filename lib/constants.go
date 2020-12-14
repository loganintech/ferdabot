package ferdabot

// FerdaMessages
var (
	// Success
	NewFerda         = FerdaSuccess("Thanks for the new ferda <@!%s>", "Thanked %s for the ferda.")
	NotFerdaMessage  = FerdaSuccess("<@!%s> is not ferda.", "%s - %s is not ferda.")
	GetFerda         = FerdaSuccess("<@!%s> was ferda on %s for %s.", "Sent a retrieved ferda Message %s.")
	EchoSuccess      = FerdaSuccess("%s", "Echo'd ferda Message")
	GetDetailedFerda = FerdaSuccess("[%d - %s] <@!%s> was ferda on %s for %s added by <@!%d>.", "Sent a detailed ferda Message %s.")
	DeletedFerda     = FerdaSuccess("Deleted a ferda successfully <@!%s>. ID: [%s]", "Deleted a ferda with Id %d.")
	HelpBody         = FerdaSuccess("> [%s] %s", "Help Message body.")
	DiceBody         = FerdaSuccess("Rolling [%dd%d%s]: %d", "Rolling [%dd%d%s]: %d")
	ChoiceResult     = FerdaSuccess("The choice is: %s", "Chose %s from random array %v")
	ReminderAdded    = FerdaSuccess("You will be reminded to: %s on %v", "New Reminder Saved")

	// Failure
	MentionMissing     = FerdaFailure("You must ping an associated user. Ex: `+ferda @Logan is a great guy` or `?ferda @Logan`", "Failed to find a username in Message %s.")
	DBInsertErr        = FerdaFailure("Database error occurred, please contact Logan.", "Error inserting into the DB %+v")
	DBDeleteErr        = FerdaFailure("Database error occurred, please contact Logan.", "Error deleting from the DB %+v.")
	UserNotFoundErr    = FerdaFailure("That user was not found.", "Couldn't load user from discord: %s")
	DBGetErr           = FerdaFailure("Database error occurred, please contact Logan.", "Error getting from table %+v")
	EchoFailure        = FerdaFailure("You shouldn't be trying to send that you dumb fuck.", "User: %s, ID: %s tried to send %s")
	MissingID          = FerdaFailure("You must include a ferda Id to delete `-ferda [Id]`", "Not given an ID of ferda to delete.")
	NotEnoughArguments = FerdaFailure("Not enough arguments, need %d minimum.", "Not enough arguments supplied to command %s.")
	BadDiceFormat      = FerdaFailure("Bad dice format. Please use the form NdN (1d6 for example).", "Bad dice format %s")
	BadReminderFormat  = FerdaFailure("Bad reminder format. Please use the form N[YMDhms] (3Y2M3D9h4m1s to mean 3 years, 2 months, 3 days, 9 hours, 4 minutes, and 1 second).", "Bad reminder format %s")
	BadIDFormat        = FerdaFailure("ID was not in the correct format: %s", "ID was not in the correct format: %s")

	// LogOnly Fail
	AddRouteFailed             = FerdaLogOnly("Adding route failed, %s already exists").SetFail()
	RouteNotFound              = FerdaLogOnly("Route %s was not found").SetFail()
	SpotifyCreatePlaylistError = FerdaLogOnly("Spotify Playlist Creation Failed [%v]").SetFail()
	SpotifyAddToPlaylistError  = FerdaLogOnly("Spotify Tracks failed to Add to Playlist [%v]").SetFail()
	SpotifyUserNotFound        = FerdaLogOnly("Current Spotify user Not Set [%v]").SetFail()
	ParamNotFound              = FerdaLogOnly("ConfigEntry was not found: %s").SetFail()
	CantCreateUserChannel      = FerdaLogOnly("Can't Create User Channel: [%s] %v").SetFail()
	CantSendUserMessage        = FerdaLogOnly("Can't Send User Message: [%s] %v").SetFail()

	// LogOnly Success
	SpotifySongAdded = FerdaLogOnly("Song [%s] Added to Playlist: %s").SetSuccess()
	AddRouteSuccess  = FerdaLogOnly("Adding route complete: %s").SetSuccess()

	// Pre-Finalized Success
	HelpHeader          = FerdaSuccess("Here's the list of available commands:", "Help Message sent.").Finalize()
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

	DontLog = FerdaActionBuilder{DontLog: true}.Finalize()
)
