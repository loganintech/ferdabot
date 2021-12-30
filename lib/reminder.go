package ferdabot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeRegex = regexp.MustCompile("([0-9]+[YMdhms])")

type Reminder struct {
	ID        int64     `db:"id"`
	CreatorID string    `db:"creatorid"`
	Timestamp time.Time `db:"time"`
	Message   string    `db:"message"`
}

func (b *Bot) processDeleteReminder(authorID string, trimmedText string) FerdaAction {
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

	if reminder, action := b.getReminder(foundID); !action.Success() {
		return action
	} else if reminder.CreatorID != authorID {
		return CantDeleteOthersReminder.RenderLogText(authorID, reminder, reminder.CreatorID).Finalize()
	}

	// Delete the ferda
	deleteAction := b.deleteReminder(foundID)
	if !deleteAction.Success() {
		return deleteAction
	}

	// return DeleteFerda success Message
	return DeletedItem.RenderDiscordText("reminder", authorID, foundID).RenderLogText(foundID).Finalize()
}

func (b *Bot) processListReminders(authorID string) FerdaAction {
	reminders, act := b.getReminders(authorID)
	if !act.Success() {
		return act
	}

	reminderMsg := ReminderHeader
	for _, reminder := range reminders {
		reminderMsg = reminderMsg.CombineActions(ReminderBody.RenderDiscordText(reminder.ID, reminder.Message, reminder.Timestamp.Format("Mon, January 02, 2006 at 03:04:05 PM")).Finalize())
	}

	if len(reminders) == 0 {
		reminderMsg = NoRemindersFound.RenderLogText(authorID).Finalize()
	}

	userChan, err := b.discord.UserChannelCreate(authorID)
	if err != nil {
		return CantCreateUserChannel.RenderLogText(authorID, err).Finalize()
	}

	_, msgErr := b.discord.ChannelMessageSend(userChan.ID, reminderMsg.DiscordText)
	if msgErr != nil {
		return CantSendUserMessage.RenderLogText(authorID, msgErr).Finalize()
	}
	return CheckYourDMs
}

func (b *Bot) processNewReminder(authorID string, trimmedText string) FerdaAction {
	args := strings.Split(trimmedText, " ")

	if len(args) < 2 {
		return NotEnoughArguments.RenderDiscordText(2).Finalize()
	}

	message := strings.Join(args[1:], " ")
	found := timeRegex.FindAllStringSubmatch(args[0], -1)
	if found == nil {
		return BadReminderFormat.RenderLogText(args[0]).Finalize()
	}

	then := time.Now()
	for _, unit := range found {
		match := unit[0]
		val, _ := strconv.Atoi(match[:len(match)-1])
		switch string(match[len(match)-1]) {
		case "Y":
			now := time.Now()
			future := time.Date(now.Year()+val, now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
			yearDiff := future.Sub(now)
			then = then.Add(yearDiff)
		case "M":
			now := time.Now()
			year, mnth := yearMonthOffset(now.Month(), val)
			future := time.Date(now.Year()+year, mnth, now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
			monthDiff := future.Sub(now)
			then = then.Add(monthDiff)
		case "d":
			now := time.Now()
			future := time.Date(now.Year(), now.Month(), now.Day()+val, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
			dayDiff := future.Sub(now)
			then = then.Add(dayDiff)
		case "h":
			fallthrough
		case "m":
			fallthrough
		case "s":
			formatted := fmt.Sprintf("%d%s", val, string(match[len(match)-1]))
			hourDiff, _ := time.ParseDuration(formatted)
			then = then.Add(hourDiff)
		}
	}

	if action := b.newReminder(then, authorID, message); !action.Success() {
		return action
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	return ReminderAdded.RenderDiscordText(message, then.In(loc).Format("Monday, January 2, 2006 at 3:04:05 PM")).Finalize()
}

func yearMonthOffset(sourceMonth time.Month, numMoreMonths int) (int, time.Month) {
	monthCounter := int(sourceMonth)
	newYears := 0
	for i := 0; i < numMoreMonths; i++ {
		if monthCounter == 12 {
			monthCounter = 0
			newYears++
		}
		monthCounter++
	}

	return newYears, monthFromInt(monthCounter)
}

func monthFromInt(month int) time.Month {
	switch month {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	case 12:
		return time.December
	}

	return time.January
}

func (b *Bot) reminderLoop() FerdaAction {
	fiveSeconds, _ := time.ParseDuration("5s")
	for {
		reminders, action := b.getOverdueReminders()
		if !action.Success() {
			b.ProcessFerdaAction(action, nil, nil)
		}

		var itemsToDelete []int64
		for _, reminder := range reminders {
			userChan, err := b.discord.UserChannelCreate(reminder.CreatorID)
			if err != nil {
				b.ProcessFerdaAction(CantCreateUserChannel.RenderLogText(reminder.CreatorID, err).Finalize(), nil, nil)
				continue
			}

			_, msgErr := b.discord.ChannelMessageSend(userChan.ID, reminder.Message)
			if msgErr != nil {
				b.ProcessFerdaAction(CantSendUserMessage.RenderLogText(reminder.CreatorID, msgErr).Finalize(), nil, nil)
				continue
			}
			itemsToDelete = append(itemsToDelete, reminder.ID)
		}

		if act := b.deleteOverdueReminders(itemsToDelete); !act.Success() {
			b.ProcessFerdaAction(act, nil, nil)
		}

		time.Sleep(fiveSeconds)
	}
}
