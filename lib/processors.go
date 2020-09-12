package ferdabot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

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

func processHelp(_ *discordgo.Session, _ *discordgo.MessageCreate, _ string) FerdaAction {
	return HelpMessage
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
	res, dbErr := b.db.NamedExec(
		`DELETE FROM ferda WHERE id = :ferdaid`,
		map[string]interface{}{
			"ferdaid": foundID,
		},
	)
	if dbErr != nil {
		return DBDeleteErr.RenderLogText(dbErr).Finalize()
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return NoRowDBErr.Finalize()
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
		ferdaDetails = ferdaDetails.AppendAction(ferdaAction)
	}

	return ferdaDetails
}
