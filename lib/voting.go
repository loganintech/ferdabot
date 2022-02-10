package ferdabot

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
)

func (b *Bot) formatVoteEmbed(title string, subject string, lines []Line, castVotes map[string][]*CastVote, closed bool) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	embed := &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Link/Topic",
				Value: subject,
			},
		},
	}

	if closed {
		embed.Description = "Vote is closed. No more votes may be cast."
		embed.Title = fmt.Sprintf("✅ %s", embed.Title)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Status: ✅ Completed ✅",
		}
		embed.Color = MODBOT_GREEN

	} else {
		embed.Description = "Vote using the reactions below:"
		embed.Title = fmt.Sprintf("⏳ %s", embed.Title)
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Status: ⏳ Ongoing ⏳",
		}
		embed.Color = MODBOT_WARNING
	}

	var rows []discordgo.MessageComponent
	reactionValue := &discordgo.MessageEmbedField{
		Name:   "Reactions",
		Value:  "",
		Inline: true,
	}

	countValue := &discordgo.MessageEmbedField{
		Name:   "Votes",
		Value:  "",
		Inline: true,
	}

	// We need to break the lines of buttons into rows of 5 at max
	for i := 0; i < (len(lines)/5)+1; i++ {
		var buttons []discordgo.MessageComponent

		// Loop through this row of buttons. Exit the loop if we hit 5 items OR if we exceed the max lines
		for x := 5 * i; (x < (i+1)*5) && x < len(lines); x++ {
			line := lines[x]

			buttons = append(buttons, &discordgo.Button{
				Style:    discordgo.SecondaryButton,
				CustomID: line.EmojiName,
				Emoji: discordgo.ComponentEmoji{
					Name: line.EmojiName,
					ID:   line.EmojiID,
				},
			})

			if x != 0 {
				reactionValue.Value += "\n"
				countValue.Value += "\n"
			}

			if line.EmojiID == "" {
				reactionValue.Value += fmt.Sprintf("%s %s", line.EmojiName, line.LineValue)
			} else {
				reactionValue.Value += fmt.Sprintf("<:%s:%s> %s", line.EmojiName, line.EmojiID, line.LineValue)
			}

			if castedVotes, ok := castVotes[line.EmojiName]; ok {
				countValue.Value += strconv.Itoa(len(castedVotes))
			} else {
				countValue.Value += "0"
			}
		}

		rows = append(rows, discordgo.ActionsRow{Components: buttons})
	}

	reactionValue.Value += "\n✅ Close this Vote (creator only)"
	embed.Fields = append(embed.Fields, reactionValue, countValue)

	if closed {
		var selectedVote string
		var voteCount int
		for key, result := range castVotes {
			if len(result) > voteCount {
				voteCount = len(result)
				selectedVote = key
			}
		}

		if selectedVote == "" {
			selectedVote = "No option selected"
		}

		var selectedLine *Line
		for _, line := range lines {
			if selectedVote == line.EmojiName {
				selectedLine = &line
				break
			}
		}

		if selectedLine != nil {
			if selectedLine.EmojiID == "" {
				selectedVote = fmt.Sprintf("%s %s", selectedLine.EmojiName, selectedLine.LineValue)
			} else {
				selectedVote = fmt.Sprintf("<:%s:%s> %s", selectedLine.EmojiName, selectedLine.EmojiID, selectedLine.LineValue)
			}
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Winning result",
			Value: selectedVote,
		})
	}

	// If we need a new row for this element, add it
	if len(lines)%5 == 0 {
		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Style:    discordgo.SuccessButton,
					CustomID: "close_vote",
					Emoji: discordgo.ComponentEmoji{
						Name: "✅",
					},
				},
			},
		})
	} else {
		// Otherwise, add to the last row

		lastRow := rows[len(rows)-1]
		lastRowCasted := lastRow.(discordgo.ActionsRow)
		lastRowCasted.Components = append(lastRowCasted.Components, &discordgo.Button{
			Style:    discordgo.SuccessButton,
			CustomID: "close_vote",
			Emoji: discordgo.ComponentEmoji{
				Name: "✅",
			},
		})

		rows[len(rows)-1] = lastRowCasted
	}

	if closed {
		rows = []discordgo.MessageComponent{}
	}

	return embed, rows
}

type entry struct {
	line  Line
	index int
}
type entryList []entry

func (e entryList) Len() int           { return len(e) }
func (e entryList) Less(i, j int) bool { return e[i].index < e[j].index }
func (e entryList) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func getLines(sentence string) ([]Line, error) {

	stringIndex := reg.FindAllStringIndex(sentence, -1)
	submatches := reg.FindAllStringSubmatch(sentence, -1)

	entries := entryList{}

	replacedSentence := sentence
	foundEmojis := emoji.FindAll(sentence)
	for _, foundEmoji := range foundEmojis {
		found, ok := foundEmoji.Match.(emoji.Emoji)
		if !ok {
			continue
		}

		if len(foundEmoji.Locations) > 1 {
			return nil, errors.New("duplicate emoji found")
		}

		entries = append(entries, entry{
			line: Line{
				EmojiName: found.Value,
			},
			index: foundEmoji.Locations[0][0],
		})

		replacedSentence = strings.ReplaceAll(replacedSentence, found.Value, "|")
	}

	var lines []Line
	for i, match := range submatches {
		if len(match) < 2 {
			continue
		}

		if len(stringIndex[i]) > 2 {
			return nil, errors.New("duplicate emoji found")
		}

		en := entry{
			line: Line{
				EmojiName: match[1],
				EmojiID:   match[2],
			},
			index: stringIndex[i][1],
		}

		entries = append(entries, en)

		replacedSentence = strings.ReplaceAll(replacedSentence, match[0], "|")
	}

	sort.Sort(entries)

	for _, e := range entries {
		lines = append(lines, e.line)
	}

	justValues := strings.Split(replacedSentence, "|")
	for i, val := range justValues {
		if i == 0 || val == "" {
			continue
		}
		lines[i-1].LineValue = strings.TrimSpace(val)
	}

	return lines, nil
}
