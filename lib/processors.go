package ferdabot

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// processEcho processes an echo command
func (b *Bot) processEcho(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Make sure they aren't trying to echo a banned phrase
	for _, phrase := range BannedEchoPhrases {
		// If it does have one, make fun of them
		if strings.Contains(trimmedText, phrase) {
			return EchoFailure.RenderLogText(m.Author.Username, m.Author.ID, trimmedText).Finalize()
		}
	}

	// Echo the result back
	return EchoSuccess.RenderDiscordText(trimmedText).Finalize()
}

// processNewFerda processes a new ferda Message
func (b *Bot) processNewFerda(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Split the string
	split := strings.Split(trimmedText, " ")
	// get our first userid
	foundString := b.getUserFromFirstWord(trimmedText)

	// if the found userid is empty return metion missing error
	if foundString == "" {
		return MentionMissing.RenderLogText(m.Author.Username).Finalize()
	}

	// Insert a new ferda entry
	ferdaAction := b.insertFerdaEntry(foundString, strings.Join(split[1:], " "), m.Author.ID)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	return NewFerda.RenderDiscordText(m.Author.ID).RenderLogText(m.Author.Username).Finalize()
}

// processGetFerda processes a get ferda Message
func (b *Bot) processGetFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Get the found user
	foundUser := b.getUserFromFirstWord(trimmedText)
	// Load extra details about them
	user, err := s.User(foundUser)
	// Return discord error if they weren't found
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	// Get the entry
	ferdaEntry, ferdaAction := b.getFerdaEntry(foundUser, user.ID, user.Username)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	// And return the GetFerda Message
	return GetFerda.RenderDiscordText(user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006"), ferdaEntry.Reason).RenderLogText(err).Finalize()
}

// processDetailedGetFerda processes a detailed getferda Message
func (b *Bot) processDetailedGetFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Get the user
	foundUser := b.getUserFromFirstWord(trimmedText)
	// Load more info from discord
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	// Get the ferda entry
	ferdaEntry, ferdaAction := b.getFerdaEntry(foundUser, user.ID, user.Username)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	// And render like a GetFerda Message, but with more details
	return GetDetailedFerda.RenderDiscordText(ferdaEntry.ID, user.ID, user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006 at 15:04:05 -0700"), ferdaEntry.Reason, ferdaEntry.CreatorID).RenderLogText(err).Finalize()
}

// processDeleteFerda processes a remove ferda Message
func (b *Bot) processDeleteFerda(_ *discordgo.Session, m *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Split into args
	split := strings.Split(trimmedText, " ")
	// And complain if we're missing an arg
	if len(split) < 1 {
		return MissingID.Finalize()
	}
	foundID := split[0]

	if _, err := strconv.Atoi(foundID); err != nil {
		return BadIDFormat.RenderLogText(foundID).RenderLogText(foundID).Finalize()
	}

	// Delete the ferda
	deleteAction := b.deleteFerda(foundID)
	if !deleteAction.Success() {
		return deleteAction
	}

	// return DeleteFerda success Message
	return DeletedItem.RenderDiscordText("ferda", m.Author.ID, foundID).RenderLogText(foundID).Finalize()
}

// processSearchFerda searches for a ferda given some text
func (b *Bot) processSearchFerda(s *discordgo.Session, _ *discordgo.MessageCreate, trimmedText string) FerdaAction {
	// Ensure there's at least a user and a search string
	data := strings.Split(trimmedText, " ")
	// If there isn't enough args, return NotEnoughArguments FerdaAction
	if len(data) < 2 {
		return NotEnoughArguments.RenderDiscordText(2).RenderLogText("ferdasearch").Finalize()
	}
	// find a username
	foundUser := b.getUserFromFirstWord(data[0])
	// And get them from discord
	user, err := s.User(foundUser)
	if err != nil {
		return UserNotFoundErr.RenderLogText(err).Finalize()
	}

	// Set the data to the search string only
	data = data[1:]
	// And join the body text
	bodyText := strings.Join(data, " ")

	// Then search for the ferda by text
	ferdaEntries, ferdaAction := b.ferdaSearch(foundUser, user.ID, user.Username, bodyText)
	if !ferdaAction.Success() {
		return ferdaAction
	}

	// And get our FerdaFound Message
	ferdaDetails := MultipleFerdasFound.Finalize()
	// Loop over every ferda we found
	for _, ferdaEntry := range ferdaEntries {
		// Create a ferda action with their details
		ferdaAction := GetDetailedFerda.RenderDiscordText(ferdaEntry.ID, user.ID, user.ID, ferdaEntry.When.Format("Mon, Jan _2 2006 at 15:04:05 -0700"), ferdaEntry.Reason, ferdaEntry.CreatorID).RenderLogText(err).Finalize()
		// And combine them into the FerdaFound Message
		ferdaDetails = ferdaDetails.CombineActions(ferdaAction)
	}

	// And return the ferdas we found
	return ferdaDetails
}

// processHelp returns a help Message from the treeRouter
func (b *Bot) processHelp(_ *discordgo.Session, _ *discordgo.MessageCreate, _ string) FerdaAction {
	// Create a HelpHeader msg to combine
	buildMsg := HelpHeader
	action := b.treeRouter.GetHelpActions()
	if action == nil {
		return RouteNotFound.Finalize()
	}
	// Get all of the results, split them into lines
	newDiscordText := strings.Split(action.DiscordText, "\n")
	// Sort them alphabetically (starting after the first one)
	sort.Strings(newDiscordText[1:])
	// And re-join the Message
	action.DiscordText = strings.Join(newDiscordText, "\n")
	return buildMsg.CombineActions(*action)
}

func (b *Bot) processPing(_ *discordgo.Session, m *discordgo.MessageCreate, _ string) FerdaAction {
	then, err := m.Timestamp.Parse()
	if err != nil {
		return TimeParseFailed.RenderLogText(m.Timestamp, err).Finalize()
	}
	now := time.Now()
	diff := now.Sub(then).Milliseconds()
	return Pong.RenderLogText(diff).RenderDiscordText(diff).Finalize()
}
