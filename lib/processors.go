package ferdabot

import (
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var diceRegex = regexp.MustCompile("([0-9]*)[dD]([0-9]*)")

func (b *Bot) processEcho(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	for _, phrase := range BannedEchoPhrases {
		if strings.Contains(trimmedText, phrase) {
			return EchoFailure.RenderLogText(m.Author.Username, m.Author.ID, trimmedText).Finalize()
		}
	}

	return EchoSuccess.RenderDiscordText(trimmedText).Finalize()
}

func (b *Bot) processNewFerda(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	split := strings.Split(trimmedText, " ")
	foundString := b.getUserFromFirstWord(trimmedText)

	if foundString == "" {
		return MentionMissing.RenderLogText(m.Author.Username).Finalize()
	}

	ferdaAction := b.insertFerdaEntry(foundString, strings.Join(split[1:], " "), m.Author.ID)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	return NewFerda.RenderDiscordText(m.Author.ID).RenderLogText(m.Author.Username).Finalize()
}

func (b *Bot) processGetFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	foundUser := b.getUserFromFirstWord(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	ferdaEntry, ferdaAction := b.getFerdaEntry(foundUser, user.ID, user.Username)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	return GetFerda.RenderDiscordText(user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006"), ferdaEntry.Reason).RenderLogText(err).Finalize()
}

func (b *Bot) processDetailedGetFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	foundUser := b.getUserFromFirstWord(trimmedText)
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	ferdaEntry, ferdaAction := b.getFerdaEntry(foundUser, user.ID, user.Username)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	return GetDetailedFerda.RenderDiscordText(ferdaEntry.ID, user.ID, user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006 at 15:04:05 -0700"), ferdaEntry.Reason, ferdaEntry.CreatorID).RenderLogText(err).Finalize()
}

func (b *Bot) processRemoveFerda(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	split := strings.Split(trimmedText, " ")
	if len(split) < 1 {
		return MissingID.Finalize()
	}
	foundID := split[0]

	deleteAction := b.deleteFerda(foundID)
	if !deleteAction.Success() {
		return deleteAction
	}

	return DeletedFerda.RenderDiscordText(m.Author.ID, foundID).RenderLogText(foundID).Finalize()
}

func (b *Bot) processSearchFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	data := strings.Split(trimmedText, " ")
	if len(data) < 2 {
		return NotEnoughArguments.RenderDiscordText(2).RenderLogText("ferdasearch").Finalize()
	}
	foundUser := b.getUserFromFirstWord(data[0])
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	data = data[1:]
	bodyText := strings.Join(data, " ")

	ferdaEntries, ferdaAction := b.ferdaSearch(foundUser, user.ID, user.Username, bodyText)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	ferdaDetails := MultipleFerdasFound.Finalize()
	for _, ferdaEntry := range ferdaEntries {
		ferdaAction := GetDetailedFerda.RenderDiscordText(ferdaEntry.ID, user.ID, user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006 at 15:04:05 -0700"), ferdaEntry.Reason, ferdaEntry.CreatorID).RenderLogText(err).Finalize()
		ferdaDetails = ferdaDetails.CombineActions(ferdaAction)
	}

	return ferdaDetails
}

func (b *Bot) processHelp(_ *discordgo.Session, _ *discordgo.MessageCreate, _ string) FerdaAction {
	buildMsg := HelpHeader
	action := b.treeRouter.GetHelpActions()
	if action == nil {
		return RouteNotFound.Finalize()
	}
	newDiscordText := strings.Split(action.DiscordText, "\n")
	sort.Strings(newDiscordText[1:])
	action.DiscordText = strings.Join(newDiscordText, "\n")
	return buildMsg.CombineActions(*action)
}

func (b *Bot) processDice(_ *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	args := strings.Split(trimmedText, " ")
	if len(args) < 1 {
		return NotEnoughArguments.RenderDiscordText(1).Finalize()
	}

	diceMsg := DiceHeader
	for _, arg := range args {
		newMsg := processDice(arg)
		if newMsg.Success() {
			diceMsg = diceMsg.CombineActions(newMsg)
		} else {
			return newMsg
		}
	}

	return diceMsg
}

func processDice(diceText string) FerdaAction {
	rand.Seed(time.Now().UnixNano())
	args := strings.Split(diceText, " ")
	if len(args) < 1 {
		return NotEnoughArguments.RenderDiscordText(1).RenderLogText("diceroll").Finalize()
	}

	found := diceRegex.FindAllStringSubmatch(diceText, -1)
	if len(found) == 0 {
		return ImproperDiceFormat.RenderLogText(diceText).Finalize()
	}

	var diceBody *FerdaAction = nil
	for _, match := range found {
		if len(match) < 3 {
			return ImproperDiceFormat.RenderLogText(match).Finalize()
		}

		count, _ := strconv.Atoi(match[1])
		sides, _ := strconv.Atoi(match[2])

		if count > 20 {
			return TooManyDiceToRoll
		}

		for i := 0; i < count; i++ {
			val := rand.Int63n(int64(sides)) + 1
			newBody := DiceBody.RenderDiscordText(count, sides, val).RenderLogText(count, sides, val).Finalize()
			if diceBody == nil {
				diceBody = &newBody
			} else {
				newBody = diceBody.CombineActions(newBody)
				diceBody = &newBody
			}
		}
	}

	if diceBody == nil {
		return ImproperDiceFormat.RenderLogText(diceText).Finalize()
	}

	return *diceBody
}
