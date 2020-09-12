package ferdabot

import "fmt"

type FerdaAction struct {
	Success           bool   `json:"success"`
	DiscordText       string `json:"discord_text"`
	DiscordTextFormat string
	LogText           string `json:"log_text"`
	LogTextFormat     string
}

func FerdaSuccess(dText, lText string) FerdaAction {
	return newFerdaAction(dText, lText).SetSuccess()
}

func FerdaFailure(dText, lText string) FerdaAction {
	return newFerdaAction(dText, lText).SetFail()
}

func newFerdaAction(dText, lText string) FerdaAction {
	return FerdaAction{
		Success:           false,
		DiscordTextFormat: dText,
		LogTextFormat:     lText,
	}
}

func (a FerdaAction) SetSuccess() FerdaAction {
	a.Success = true
	return a
}

func (a FerdaAction) SetFail() FerdaAction {
	a.Success = false
	return a
}

func (a FerdaAction) RenderDiscordText(d ...interface{}) FerdaAction {
	a.DiscordText = fmt.Sprintf(a.DiscordTextFormat, d...)
	return a
}

func (a FerdaAction) RenderLogText(l ...interface{}) FerdaAction {
	a.LogText = fmt.Sprintf(a.LogTextFormat, l...)
	return a
}
