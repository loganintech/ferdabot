package ferdabot

import (
	"github.com/jmoiron/sqlx"
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
	// Find a ferda where the ID is supplied
	res, dbErr := b.db.NamedExec(
		`DELETE FROM ferda WHERE id = :ferdaid`,
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

// newReminder sets a new reminder
func (b *Bot) newReminder(remind time.Time, userid string, message string) FerdaAction {
	// Named Execution insert to the ferda table
	res, dbErr := b.db.NamedExec(
		`INSERT INTO reminder (creatorid, time, message) VALUES (:creatorid, :time, :message)`,
		map[string]interface{}{
			"creatorid": userid,
			"time":      remind,
			"message":   message,
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

// getOverdueReminders returns a list of reminders that should be reminded, then deleted
func (b *Bot) getOverdueReminders() ([]Reminder, FerdaAction) {
	var reminders []Reminder
	dbErr := b.db.Select(
		&reminders,
		`SELECT * FROM reminder WHERE time < NOW()`,
	)
	// If the dbErr isn't nil
	if dbErr != nil && dbErr.Error() != "sql: no rows in result set" {
		// Return the DBGetErr from the dbErr
		return reminders, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return the found entry and DBSuccess FerdaAction
	return reminders, DBSuccess.Finalize()
}

// delete returns a list of reminders that should be reminded, then deleted
func (b *Bot) deleteOverdueReminders(ids []int64) FerdaAction {
	if len(ids) == 0 {
		return DBSuccess.Finalize()
	}
	query, args, _ := sqlx.In("DELETE FROM reminder WHERE id IN (?);", ids)

	query = b.db.Rebind(query)
	res, dbErr := b.db.Exec(query, args...)
	if dbErr != nil {
		return DBDeleteErr.RenderLogText(dbErr).Finalize()
	}

	// If no rows are affected, return no rows affected error
	count, rowErr := res.RowsAffected()
	if rowErr != nil {
		return DBDeleteErr.RenderLogText(rowErr).Finalize()
	}
	if count != int64(len(ids)) {
		return NoRowDBErr.Finalize()
	}

	return DBSuccess.Finalize()
}

// getReminders loads a list of reminders based on the user
func (b *Bot) getReminders(userid string) ([]Reminder, FerdaAction) {
	var reminders []Reminder
	dbErr := b.db.Select(&reminders, `SELECT * FROM reminder WHERE creatorid = $1`, userid)
	// If the dbErr isn't nil
	if dbErr != nil {
		// And the error was a not found error
		if dbErr.Error() == "sql: no rows in result set" {
			// Return that the user isn't ferda
			return reminders, NoRemindersFound.RenderLogText(userid).Finalize()
		}
		// Return the DBGetErr from the dbErr
		return reminders, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return the found reminders and DBSuccess FerdaAction
	return reminders, DBSuccess
}

// getReminder loads a reminder based on its ID
func (b *Bot) getReminder(foundID string) (Reminder, FerdaAction) {
	var reminder Reminder
	dbErr := b.db.Get(&reminder, `SELECT * FROM reminder WHERE id = $1`, foundID)
	// If the dbErr isn't nil
	if dbErr != nil {
		// And the error was a not found error
		if dbErr.Error() == "sql: no rows in result set" {
			// Return that the user isn't ferda
			return reminder, NoRemindersFoundByID.RenderLogText(foundID).Finalize()
		}
		// Return the DBGetErr from the dbErr
		return reminder, DBGetErr.RenderLogText(dbErr).Finalize()
	}

	// Return the found reminder and DBSuccess FerdaAction
	return reminder, DBSuccess
}

// deleteReminder deletes a reminder
func (b *Bot) deleteReminder(reminderID string) FerdaAction {
	// Delete a reminder based on its ID
	res, dbErr := b.db.NamedExec(
		`DELETE FROM reminder WHERE id = :reminderid`,
		map[string]interface{}{
			"reminderid": reminderID,
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
