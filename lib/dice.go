package ferdabot

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// diceRegex matches the 1d6 format for dice rolling
var diceRegex = regexp.MustCompile("([0-9]*)[dD]([0-9]*)[+]?([0-9]+)?")

// processDice processes the dice roll command
func (b *Bot) processDice(_ *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Split the args
	args := strings.Split(trimmedText, " ")
	// Validate the arguments
	if len(args) < 1 {
		return NotEnoughArguments.RenderDiscordText(1).Finalize()
	}

	config, act := b.getConfigEntry("maxdice")
	if act.DBNotFound() {
		config.Val = "20"
	} else if !act.Success() && !act.DBNotFound() {
		return act
	}

	limit, limitParse := strconv.Atoi(config.Val)
	if limitParse != nil {
		return NumberParseError.RenderLogText(config.Val, limitParse).Finalize()
	}

	// Create a dice header
	diceMsg := DiceHeader
	// For each dice arg, get a Message for it
	for _, arg := range args {
		newMsg := processDice(arg, int64(limit))
		if newMsg.Success() {
			diceMsg = diceMsg.CombineActions(newMsg)
		} else {
			return newMsg
		}
	}

	return diceMsg
}

// processDice returns a FerdaAction based on dice rolls
func processDice(diceText string, limit int64) FerdaAction {
	// Seed the random generator
	rand.Seed(time.Now().UnixNano())
	args := strings.Split(diceText, " ")
	if len(args) < 1 {
		return NotEnoughArguments.RenderDiscordText(1).RenderLogText("diceroll").Finalize()
	}

	// Find all matches for the dice regex
	found := diceRegex.FindAllStringSubmatch(diceText, -1)
	if len(found) == 0 {
		return BadDiceFormat.RenderLogText(diceText).Finalize()
	}

	// Create a dice body
	var diceBody *FerdaAction = nil
	for _, match := range found {
		// If the found match is too short, return bad format
		if len(match) < 3 {
			return BadDiceFormat.RenderLogText(match).Finalize()
		}

		// Format the parameters
		count, _ := strconv.Atoi(match[1])
		sides, _ := strconv.Atoi(match[2])

		// See if we have to add anything
		add := 0
		addTxt := ""
		// if we have that match
		if len(match) == 4 {
			// Parse it, add to the string
			add, _ = strconv.Atoi(match[3])
			if add != 0 {
				addTxt = "+" + match[3]
			}
		}

		// Limit them to the max
		if int64(count) > limit {
			return TooManyDiceToRoll
		}

		var vals []string
		// For each dice we have to roll
		for i := 0; i < count; i++ {
			// Actually roll it
			val := rand.Int63n(int64(sides)) + 1 + int64(add)
			vals = append(vals, fmt.Sprintf("%d", val))
		}
		valStr := strings.Join(vals, ",")
		newBody := DiceBody.RenderDiscordText(count, sides, addTxt, valStr).RenderLogText(count, sides, addTxt, valStr).Finalize()
		// Create a FerdaAction with the results
		if diceBody == nil {
			diceBody = &newBody
		} else {
			newBody = diceBody.CombineActions(newBody)
			diceBody = &newBody
		}
	}

	// If we rolled no dice bodies, return bad format
	if diceBody == nil {
		return BadDiceFormat.RenderLogText(diceText).Finalize()
	}

	// Return the results
	return *diceBody
}
