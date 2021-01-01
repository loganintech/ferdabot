package ferdabot

import "fmt"

// FerdaAction represents a finalized Action
type FerdaAction struct {
	DiscordText string `json:"discord_text"`
	LogText     string `json:"log_text"`

	// flags
	DontLog    bool
	LogOnly    bool
	success    bool
	dbNotFound bool
}

// FerdaActionBuilder lets you build a ferda action
type FerdaActionBuilder struct {
	// text data
	discordText       string
	discordTextFormat string
	logText           string
	logTextFormat     string

	// flags
	DontLog    bool
	Success    bool
	LogOnly    bool
	DBNotFound bool
}

// FerdaSuccess returns a success FerdaAction
func FerdaSuccess(dText, lText string) FerdaActionBuilder {
	return newFerdaAction(dText, lText).SetSuccess()
}

// FerdaFailure returns a fail FerdaAction
func FerdaFailure(dText, lText string) FerdaActionBuilder {
	return newFerdaAction(dText, lText).SetFail()
}

// FerdaLogOnly returns a log only FerdaAction
func FerdaLogOnly(lText string) FerdaActionBuilder {
	a := newFerdaAction("", lText)
	a.LogOnly = true
	return a
}

// newFerdaAction lets you set a basic FerdaAction
func newFerdaAction(dText, lText string) FerdaActionBuilder {
	return FerdaActionBuilder{
		Success:           false,
		discordTextFormat: dText,
		logTextFormat:     lText,
		LogOnly:           false,
		DontLog:           false,
	}
}

// SetSuccess returns an identical FerdaAction with the success set to true
func (a FerdaActionBuilder) SetSuccess() FerdaActionBuilder {
	a.Success = true
	return a
}

// SetFail returns an identical FerdaAction with the success set to false
func (a FerdaActionBuilder) SetFail() FerdaActionBuilder {
	a.Success = false
	return a
}

// SetFail returns an identical FerdaAction with the success set to false
func (a FerdaActionBuilder) SetLogOnly(val bool) FerdaActionBuilder {
	a.LogOnly = val
	return a
}

// SetDBNotFound sets whether this is a DB Not Found error
func (a FerdaActionBuilder) SetDBNotFound() FerdaActionBuilder {
	a.DBNotFound = true
	return a
}

// RenderDiscordText formats the discordTextFormat to discordText
func (a FerdaActionBuilder) RenderDiscordText(d ...interface{}) FerdaActionBuilder {
	a.discordText = fmt.Sprintf(a.discordTextFormat, d...)
	return a
}

// RenderLogText formats the logTextFormat to logText
func (a FerdaActionBuilder) RenderLogText(l ...interface{}) FerdaActionBuilder {
	a.logText = fmt.Sprintf(a.logTextFormat, l...)
	return a
}

// Finalize consumes the FerdaActionBuilder to return a finalized FerdaAction
func (a FerdaActionBuilder) Finalize() FerdaAction {
	if a.discordText == "" {
		a.discordText = a.discordTextFormat
	}
	if a.logText == "" {
		a.logText = a.logTextFormat
	}
	return FerdaAction{
		DiscordText: a.discordText,
		DontLog:     a.DontLog,
		LogOnly:     a.LogOnly,
		LogText:     a.logText,
		dbNotFound:  a.DBNotFound,
		success:     a.Success,
	}
}

// Finalize on a FerdaAction does nothing
func (a FerdaAction) Finalize() FerdaAction {
	return a
}

// Success on a FerdaAction returns whether it was a success
func (a FerdaAction) Success() bool {
	return a.success
}

// CombineActions returns a FerdaAction with the messages combined
func (a FerdaAction) CombineActions(other FerdaAction) FerdaAction {
	a.LogText = a.LogText + "\n" + other.LogText
	a.LogOnly = a.LogOnly && other.LogOnly
	a.DiscordText = a.DiscordText + "\n" + other.DiscordText
	a.DontLog = a.DontLog && other.DontLog
	a.dbNotFound = a.dbNotFound || other.dbNotFound
	return a
}

// Equals returns the similarity between two FerdaAction's
func (a FerdaAction) Equals(other FerdaAction) bool {
	return a.DiscordText == other.DiscordText && a.LogText == other.LogText && a.LogOnly == other.LogOnly
}

// DBNotFound returns the similarity between two FerdaAction's
func (a FerdaAction) DBNotFound() bool {
	return a.dbNotFound
}
