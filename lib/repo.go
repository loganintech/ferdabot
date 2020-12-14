package ferdabot

import (
	"time"
)

// getFerdaEntry takes a foundUser, userID, userName to return a random FerdaEntry
func (b *Bot) getFerdaEntry(foundUser, userID, userName string) (FerdaEntry, FerdaAction) {
	ferdaEntry := FerdaEntry{}
	dbErr := b.db.Get(
		&ferdaEntry,
		`SELECT * FROM ferda WHERE userid = $1 ORDER BY RANDOM()`,
		foundUser,
	)
	// If the dbErr isn't nil
	if dbErr != nil {
		// And the error was a not found error
		if dbErr.Error() == "sql: no rows in result set" {
			// Return that the user isn't ferda
			return ferdaEntry, NotFerdaMessage.RenderDiscordText(userID).RenderLogText(userName, userID).Finalize()
		}
		// Return the DBGetErr from the dbErr
		return ferdaEntry, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return the found entry and DBSuccess FerdaAction
	return ferdaEntry, DBSuccess.Finalize()
}

// insertFerdaEntry inserts a new FerdaEntry into the database
func (b *Bot) insertFerdaEntry(foundString, reason, creatorID string) FerdaAction {
	// Named Execution insert to the ferda table
	res, dbErr := b.db.NamedExec(
		`INSERT INTO ferda (userid, time, reason, CreatorID) VALUES (:userid, :time, :reason, :CreatorID)`,
		map[string]interface{}{
			"userid": foundString,
			// Adjust for local time of bot
			"time":      time.Now().Round(time.Microsecond).Add(-(time.Hour * 7)),
			"reason":    reason,
			"CreatorID": creatorID,
		},
	)
	// If dbErr isn't nil, return the DBInsertErr log
	if dbErr != nil {
		return DBInsertErr.RenderLogText(dbErr).Finalize()
	}

	// Get the rows affected
	count, _ := res.RowsAffected()
	// If the count is 0, return NoRowDB FerdaAction
	if count == 0 {
		return NoRowDBErr
	}

	// return DBSuccess FerdaAction
	return DBSuccess
}

// ferdaSearch takes the foundUser, userID, userName, and a search string to return a list of FerdaEntries and FerdaAction
func (b *Bot) ferdaSearch(foundUser, userID, userName, searchText string) ([]FerdaEntry, FerdaAction) {
	var ferdaEntry []FerdaEntry
	// Select matchinf Ferdas
	dbErr := b.db.Select(
		&ferdaEntry,
		`SELECT * FROM ferda WHERE userid = $1 AND reason LIKE $2`,
		foundUser,
		"%"+searchText+"%",
	)
	// If we found none, return empty DB result
	if len(ferdaEntry) == 0 {
		return ferdaEntry, DBResultsEmpty.Finalize()
	}
	// if the dbErr isn't nil
	if dbErr != nil {
		// and the error is no rows found
		if dbErr.Error() == "sql: no rows in result set" {
			// return user isn't ferda
			return ferdaEntry, NotFerdaMessage.RenderDiscordText(userID).RenderLogText(userName, userID).Finalize()
		}
		// Return DBGetErr
		return ferdaEntry, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return ferdaEntries and DBSuccess
	return ferdaEntry, DBSuccess.Finalize()
}

// deleteFerda removes a ferda from the database
func (b *Bot) deleteFerda(foundID string) FerdaAction {
	// Find a ferda where the Id is supplied
	res, dbErr := b.db.NamedExec(
		`DELETE FROM ferda WHERE Id = :ferdaid`,
		map[string]interface{}{
			"ferdaid": foundID,
		},
	)
	if dbErr != nil {
		return DBDeleteErr.RenderLogText(dbErr).Finalize()
	}

	// If no rows are affected, return no rows affected error
	count, _ := res.RowsAffected()
	if count == 0 {
		return NoRowDBErr.Finalize()
	}

	return DBSuccess
}

// insertConfigEntry inserts something into the config table
func (b *Bot) insertConfigEntry(param, val string) FerdaAction {
	// Named Execution insert to the ferda table
	res, dbErr := b.db.NamedExec(
		`INSERT INTO config (key, val) VALUES (:key, :val)`,
		map[string]interface{}{
			"key": param,
			// Adjust for local time of bot
			"val": val,
		},
	)
	// If dbErr isn't nil, return the DBInsertErr log
	if dbErr != nil {
		return DBInsertErr.RenderLogText(dbErr).Finalize()
	}

	// Get the rows affected
	count, _ := res.RowsAffected()
	// If the count is 0, return NoRowDB FerdaAction
	if count == 0 {
		return NoRowDBErr
	}

	// return DBSuccess FerdaAction
	return DBSuccess
}

// getConfigEntry returns a config entry based on its param name
func (b *Bot) getConfigEntry(param string) (ConfigEntry, FerdaAction) {
	entry := ConfigEntry{}
	dbErr := b.db.Get(
		&entry,
		`SELECT * FROM config WHERE key = $1`,
		param,
	)
	// If the dbErr isn't nil
	if dbErr != nil {
		// And the error was a not found error
		if dbErr.Error() == "sql: no rows in result set" {
			// Return that the user isn't ferda
			return entry, ParamNotFound.RenderLogText(param).Finalize()
		}
		// Return the DBGetErr from the dbErr
		return entry, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return the found entry and DBSuccess FerdaAction
	return entry, DBSuccess
}
