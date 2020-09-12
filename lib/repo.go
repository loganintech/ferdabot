package ferdabot

import (
	"time"
)

func (b *Bot) getFerdaEntry(foundUser, userID, userName string) (FerdaEntry, FerdaAction) {
	ferdaEntry := FerdaEntry{}
	dbErr := b.db.Get(
		&ferdaEntry,
		`SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`,
		foundUser,
	)
	if dbErr != nil {
		if dbErr.Error() == "sql: no rows in result set" {
			return ferdaEntry, NotFerdaMessage.RenderDiscordText(userID).RenderLogText(userName, userID).Finalize()
		}
		return ferdaEntry, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	return ferdaEntry, DBSuccess.Finalize()
}

func (b *Bot) insertFerdaEntry(foundString, reason, creatorID string) FerdaAction {
	res, dbErr := b.db.NamedExec(
		`INSERT INTO ferda (userid, time, reason, creatorid) VALUES (:userid, :time, :reason, :creatorid)`,
		map[string]interface{}{
			"userid": foundString,
			// Adjust for local time of bot
			"time":      time.Now().Round(time.Microsecond).Add(-(time.Hour * 7)),
			"reason":    reason,
			"creatorid": creatorID,
		},
	)
	if dbErr != nil {
		return DBInsertErr.RenderLogText(dbErr).Finalize()
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return NoRowDBErr
	}

	return DBSuccess
}

func (b *Bot) ferdaSearch(foundUser, userID, userName, searchText string) ([]FerdaEntry, FerdaAction) {
	ferdaEntry := []FerdaEntry{}
	dbErr := b.db.Select(
		&ferdaEntry,
		`SELECT * FROM ferda WHERE userid = $1 AND reason LIKE $2`,
		foundUser,
		"%"+searchText+"%",
	)
	if len(ferdaEntry) == 0 {
		return ferdaEntry, DBResultsEmpty.Finalize()
	}
	if dbErr != nil {
		if dbErr.Error() == "sql: no rows in result set" {
			return ferdaEntry, NotFerdaMessage.RenderDiscordText(userID).RenderLogText(userName, userID).Finalize()
		}
		return ferdaEntry, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	return ferdaEntry, DBSuccess.Finalize()
}
