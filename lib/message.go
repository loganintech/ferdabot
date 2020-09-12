package ferdabot

import "fmt"

type FerdaAction struct {
	DiscordText string `json:"discord_text"`
	LogText     string `json:"log_text"`

	// flags
	LogOnly bool
	success bool
}

type FerdaActionBuilder struct {
	// text data
	discordText       string
	discordTextFormat string
	logText           string
	logTextFormat     string

	// flags
	Success bool
	LogOnly bool
}

func FerdaSuccess(dText, lText string) FerdaActionBuilder {
	return newFerdaAction(dText, lText).SetSuccess()
}

func FerdaFailure(dText, lText string) FerdaActionBuilder {
	return newFerdaAction(dText, lText).SetFail()
}

func FerdaLogOnly(lText string) FerdaActionBuilder {
	a := newFerdaAction("", lText)
	a.LogOnly = true
	return a
}

func newFerdaAction(dText, lText string) FerdaActionBuilder {
	return FerdaActionBuilder{
		Success:           false,
		discordTextFormat: dText,
		logTextFormat:     lText,
		LogOnly:           false,
	}
}

func (a FerdaActionBuilder) SetSuccess() FerdaActionBuilder {
	a.Success = true
	return a
}

func (a FerdaActionBuilder) SetFail() FerdaActionBuilder {
	a.Success = false
	return a
}

func (a FerdaActionBuilder) RenderDiscordText(d ...interface{}) FerdaActionBuilder {
	a.discordText = fmt.Sprintf(a.discordTextFormat, d...)
	return a
}

func (a FerdaActionBuilder) RenderLogText(l ...interface{}) FerdaActionBuilder {
	a.logText = fmt.Sprintf(a.logTextFormat, l...)
	return a
}

func (a FerdaActionBuilder) Finalize() FerdaAction {
	if a.discordText == "" {
		a.discordText = a.discordTextFormat
	}
	if a.logText == "" {
		a.logText = a.logTextFormat
	}
	return FerdaAction{
		DiscordText: a.discordText,
		LogText:     a.logText,
		LogOnly:     a.LogOnly,
		success:     a.Success,
	}
}

func (a FerdaAction) Finalize() FerdaAction {
	return a
}

func (a FerdaAction) Success() bool {
	return a.success
}

func (a FerdaAction) AppendAction(other FerdaAction) FerdaAction {
	a.LogText = a.LogText + "\n" + other.LogText
	a.LogOnly = a.LogOnly || other.LogOnly
	a.DiscordText = a.DiscordText + "\n" + other.DiscordText
	return a
}
