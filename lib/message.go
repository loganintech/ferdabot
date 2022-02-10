package ferdabot

import "fmt"

// Action represents a finalized Action
type Action struct {
	DiscordText string `json:"discord_text"`
	LogText     string `json:"log_text"`

	// flags
	DontLog         bool
	LogOnly         bool
	success         bool
	dbNotFound      bool
	ResponseHandled bool
}

// ActionBuilder lets you build a ferda action
type ActionBuilder struct {
	// text data
	discordText       string
	discordTextFormat string
	logText           string
	logTextFormat     string

	// flags
	DontLog         bool
	Success         bool
	LogOnly         bool
	DBNotFound      bool
	ResponseHandled bool
}

// ActionSuccess returns a success Action
func ActionSuccess(dText, lText string) ActionBuilder {
	return newFerdaAction(dText, lText).SetSuccess()
}

// ActionFailure returns a fail Action
func ActionFailure(dText, lText string) ActionBuilder {
	return newFerdaAction(dText, lText).SetFail()
}

// FerdaLogOnly returns a log only Action
func FerdaLogOnly(lText string) ActionBuilder {
	a := newFerdaAction("", lText)
	a.LogOnly = true
	return a
}

// newFerdaAction lets you set a basic Action
func newFerdaAction(dText, lText string) ActionBuilder {
	return ActionBuilder{
		Success:           false,
		discordTextFormat: dText,
		logTextFormat:     lText,
		LogOnly:           false,
		DontLog:           false,
	}
}

// SetSuccess returns an identical Action with the success set to true
func (a ActionBuilder) SetSuccess() ActionBuilder {
	a.Success = true
	return a
}

// SetFail returns an identical Action with the success set to false
func (a ActionBuilder) SetFail() ActionBuilder {
	a.Success = false
	return a
}

// SetLogOnly returns an identical Action with the log only set to the input
func (a ActionBuilder) SetLogOnly(val bool) ActionBuilder {
	a.LogOnly = val
	return a
}

// SetDBNotFound sets whether this is a DB Not Found error
func (a ActionBuilder) SetDBNotFound() ActionBuilder {
	a.DBNotFound = true
	return a
}

// RenderDiscordText formats the discordTextFormat to discordText
func (a ActionBuilder) RenderDiscordText(d ...interface{}) ActionBuilder {
	a.discordText = fmt.Sprintf(a.discordTextFormat, d...)
	return a
}

// RenderLogText formats the logTextFormat to logText
func (a ActionBuilder) RenderLogText(l ...interface{}) ActionBuilder {
	a.logText = fmt.Sprintf(a.logTextFormat, l...)
	return a
}

// Finalize consumes the ActionBuilder to return a finalized Action
func (a ActionBuilder) Finalize() Action {
	if a.discordText == "" {
		a.discordText = a.discordTextFormat
	}
	if a.logText == "" {
		a.logText = a.logTextFormat
	}
	return Action{
		DiscordText: a.discordText,
		DontLog:     a.DontLog,
		LogOnly:     a.LogOnly,
		LogText:     a.logText,
		dbNotFound:  a.DBNotFound,
		success:     a.Success,
	}
}

// Finalize on a Action does nothing
func (a Action) Finalize() Action {
	return a
}

// Success on a Action returns whether it was a success
func (a Action) Success() bool {
	return a.success
}

// CombineActions returns a Action with the messages combined
func (a Action) CombineActions(other Action) Action {
	a.LogText = a.LogText + "\n" + other.LogText
	a.LogOnly = a.LogOnly && other.LogOnly
	a.DiscordText = a.DiscordText + "\n" + other.DiscordText
	a.DontLog = a.DontLog && other.DontLog
	a.dbNotFound = a.dbNotFound || other.dbNotFound
	return a
}

// Equals returns the similarity between two Action's
func (a Action) Equals(other Action) bool {
	return a.DiscordText == other.DiscordText && a.LogText == other.LogText && a.LogOnly == other.LogOnly
}

// DBNotFound returns the similarity between two Action's
func (a Action) DBNotFound() bool {
	return a.dbNotFound
}
