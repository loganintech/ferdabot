package ferdabot

// FerdaMessages
var (
	// Success
	NewFerda         = ActionSuccess("Thanks for the new ferda <@!%s>", "Thanked %s for the ferda.")
	NotFerdaMessage  = ActionSuccess("<@!%s> is not ferda.", "%s - %s is not ferda.")
	GetFerda         = ActionSuccess("<@!%s> was ferda on %s for %s.", "Sent a retrieved ferda Message %s.")
	EchoSuccess      = ActionSuccess("%s", "Echo'd ferda Message")
	GetDetailedFerda = ActionSuccess("[%d - %s] <@!%s> was ferda on %s for %s added by <@!%d>.", "Sent a detailed ferda Message %s.")
	DeletedItem      = ActionSuccess("Deleted a %s successfully <@!%s>. ID: [%s]", "Deleted a ferda with ID %d.")
	HelpBody         = ActionSuccess("> [%s] %s", "")
	DiceBody         = ActionSuccess("Rolling [%dd%d%s]: %s", "Rolling [%dd%d%s]: %s")
	ReminderBody     = ActionSuccess("[ID: %d] Reminding you to '%s' at %s", "")
	ChoiceResult     = ActionSuccess("The choice is: %s", "Chose %s from random array %v")
	ReminderAdded    = ActionSuccess("You will be reminded to: %s on %v", "New Reminder Saved")
	Pong             = ActionSuccess("Pong! [%dms]", "Pong! [%dms]")

	// Failure
	MentionMissing           = ActionFailure("You must ping an associated user. Ex: `+ferda @Logan is a great guy` or `?ferda @Logan`", "Failed to find a username in Message %s.")
	DBInsertErr              = ActionFailure("Database error occurred, please contact Logan.", "Error inserting into the DB %+v")
	DBDeleteErr              = ActionFailure("Database error occurred, please contact Logan.", "Error deleting from the DB %+v.")
	UserNotFoundErr          = ActionFailure("That user was not found.", "Couldn't load user from discord: %s")
	DBGetErr                 = ActionFailure("Database error occurred, please contact Logan.", "Error getting from table %+v")
	EchoFailure              = ActionFailure("You shouldn't be trying to send that you dumb fuck.", "User: %s, ID: %s tried to send %s")
	MissingID                = ActionFailure("You must include a ferda ID to delete `-ferda [ID]`", "Not given an ID of ferda to delete.")
	NotEnoughArguments       = ActionFailure("Not enough arguments, need %d minimum.", "Not enough arguments supplied to command %s.")
	BadDiceFormat            = ActionFailure("Bad dice format. Please use the form NdN (1d6 for example).", "Bad dice format %s")
	BadReminderFormat        = ActionFailure("Bad reminder format. Please use the form N[YMdhms] (3Y2M3d9h4m1s to mean 3 years, 2 months, 3 days, 9 hours, 4 minutes, and 1 second).", "Bad reminder format %s")
	BadIDFormat              = ActionFailure("ID was not in the correct format: %s", "ID was not in the correct format: %s")
	NoRemindersFound         = ActionFailure("You have no pending reminders.", "No reminders found for user: %s")
	CantDeleteOthersReminder = ActionFailure("You cannot delete reminders you didn't make.", "User %s tried to delete reminder %+v from %s.")
	NumberParseError         = ActionFailure("Something was expecting a number but didn't get it, please contact Logan.", "Error Occurred parsing %s to int: %v")
	TimeParseFailed          = ActionFailure("A bad time has been attempted to parse.", "Time [%s] couldn't parse: [%v]")

	// LogOnly Fail
	AddRouteFailed        = FerdaLogOnly("Adding route failed, %s already exists").SetFail()
	RouteNotFound         = FerdaLogOnly("Route %s was not found").SetFail()
	ParamNotFound         = FerdaLogOnly("ConfigEntry was not found: %s").SetFail().SetDBNotFound()
	CantCreateUserChannel = FerdaLogOnly("Can't Create User Channel: [%s] %v").SetFail()
	CantSendUserMessage   = FerdaLogOnly("Can't Send User Message: [%s] %v").SetFail()
	NoRemindersFoundByID  = FerdaLogOnly("No reminders found for ID: %s").SetFail()
	MessageDeleteFailed   = FerdaLogOnly("Couldn't Delete Message: [%s] %v").SetFail()

	// LogOnly Success
	AddRouteSuccess = FerdaLogOnly("Adding route complete: %s").SetSuccess()

	// Pre-Finalized Success
	HelpHeader          = ActionSuccess("Here's the list of available commands:", "Help Message sent.").Finalize()
	DBSuccess           = FerdaLogOnly("DB Success").SetSuccess().Finalize()
	MultipleFerdasFound = ActionSuccess("Found the following ferdas:", "Found multiple ferdas:").Finalize()
	DiceHeader          = ActionSuccess("Rolling dice:", "Rolling Dice:").Finalize()
	ReminderHeader      = ActionSuccess("Here's your reminders:", "Reminder Message Sent").Finalize()
	OnlyPong            = ActionSuccess("Pong!", "Pong!").Finalize()
	CheckYourDMs        = ActionSuccess("A message was sent to you directly.", "").Finalize()
	//RoutesFinished      = FerdaLogOnly("Routes finished setup without error.").SetSuccess().Finalize()

	// Pre-Finalized Failure
	NoRowDBErr          = ActionFailure("Database error occurred, please contact Logan.", "No rows affected in DB.").Finalize()
	DBResultsEmpty      = ActionFailure("No entries found.", "No rows returned in result set.")
	TooManyDiceToRoll   = ActionFailure("You can't roll that many dice.", "User wanted to roll too many dice.").SetFail().Finalize()
	FromChannelNotFound = FerdaLogOnly("From Channel not Found").SetFail().Finalize()

	DontLog         = ActionBuilder{DontLog: true}.Finalize()
	ResponseHandled = ActionBuilder{ResponseHandled: true}.Finalize()
)
