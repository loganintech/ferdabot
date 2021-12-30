package ferdabot

import (
	"math/rand"
	"strings"
	"time"
)

func (b *Bot) processChoice(trimmedText string) FerdaAction {
	rand.Seed(time.Now().UnixNano())

	// Split the args
	args := strings.Split(trimmedText, "|")
	// Validate the arguments
	if len(args) < 1 {
		return NotEnoughArguments.RenderDiscordText(2).Finalize()
	}

	choiceLoc := rand.Int63n(int64(len(args)))
	return ChoiceResult.RenderDiscordText(args[choiceLoc]).RenderLogText(args[choiceLoc], args).Finalize()
}
