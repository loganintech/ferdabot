package ferdabot

import (
	"errors"
	"time"
)

type Vote struct {
	MessageID   string    `db:"message_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	CreatorID   string    `db:"creator_id"`
	CreatedAt   time.Time `db:"created_at"`
	Active      bool      `db:"active"`
	Lines       []Line
}

type Line struct {
	MessageID string `db:"message_id"`
	EmojiName string `db:"emoji_name"`
	EmojiID   string `db:"emoji_id"`
	LineValue string `db:"line_value"`
}

type CastVote struct {
	MessageID string `db:"message_id"`
	EmojiName string `db:"emoji_name"`
	AuthorID  string `db:"author_id"`
}

var NoRowsAffected = errors.New("no rows affected by DB operation")

func (b *Bot) CreateNewVote(vote *Vote) error {
	ret, dbErr := b.db.NamedExec(
		`INSERT INTO vote (message_id, title, description, creator_id, created_at, active) VALUES (:message_id, :title, :description, :creator_id, NOW(), TRUE)`,
		map[string]interface{}{
			"message_id":  vote.MessageID,
			"title":       vote.Title,
			"description": vote.Description,
			"creator_id":  vote.CreatorID,
		},
	)

	if dbErr != nil {
		return dbErr
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NoRowsAffected
	}
	return nil
}

func (b *Bot) CloseVotePost(messageID string) error {
	ret, dbErr := b.db.NamedExec(`UPDATE vote AS v SET active = FALSE WHERE v.message_id = :message_id`,
		map[string]interface{}{
			"message_id": messageID,
		})
	if dbErr != nil {
		return dbErr
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NoRowsAffected
	}
	return nil
}

func (b *Bot) GetVotePost(messageID string) (*Vote, error) {
	vote := &Vote{}
	dbErr := b.db.Get(
		vote,
		`SELECT * FROM vote WHERE message_id = $1`,
		messageID,
	)
	if dbErr != nil {
		return nil, dbErr
	}
	return vote, nil
}

func (b *Bot) AddLineToVote(line *Line) error {
	ret, dbErr := b.db.NamedExec(`INSERT INTO vote_lines (message_id, emoji_name, emoji_id, line_value) VALUES (:message_id, :emoji_name, :emoji_id, :line_value)`,
		map[string]interface{}{
			"message_id": line.MessageID,
			"emoji_name": line.EmojiName,
			"emoji_id":   line.EmojiID,
			"line_value": line.LineValue,
		})
	if dbErr != nil {
		return dbErr
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NoRowsAffected
	}
	return nil
}

func (b *Bot) GetLinesForVote(messageID string) ([]Line, error) {
	vote, err := b.GetVotePost(messageID)
	if err != nil {
		return nil, err
	}

	if vote == nil {
		return nil, errors.New("couldn't find vote")
	}

	var lines []Line
	dbErr := b.db.Select(&lines, `SELECT * FROM vote_lines WHERE message_id = $1`, messageID)
	if dbErr != nil {
		return nil, dbErr
	}

	return lines, nil
}

func (b *Bot) GetCastVotes(messageID string) ([]CastVote, error) {
	vote, err := b.GetVotePost(messageID)
	if err != nil {
		return nil, err
	}

	if vote == nil {
		return nil, errors.New("couldn't find vote")
	}

	var castedVotes []CastVote
	dbErr := b.db.Select(&castedVotes, `SELECT * FROM vote_cast WHERE message_id = $1`, messageID)
	if dbErr != nil {
		return nil, dbErr
	}

	return castedVotes, nil
}

func (b *Bot) GetCastVote(messageID, authorID string) (*CastVote, error) {
	castVote := &CastVote{}
	dbErr := b.db.Get(
		castVote,
		`SELECT * FROM vote_cast WHERE message_id = $1 AND author_id = $2`,
		messageID,
		authorID,
	)
	if dbErr != nil {
		return nil, dbErr
	}
	return castVote, nil
}

func (b *Bot) CastVote(castVote *CastVote) error {
	vote, err := b.GetVotePost(castVote.MessageID)
	if err != nil {
		return err
	}

	if vote == nil {
		return errors.New("couldn't find vote")
	}

	ret, dbErr := b.db.NamedExec(
		`INSERT INTO vote_cast AS v (message_id, emoji_name, author_id)
 			  VALUES (:message_id, :emoji_name, :author_id)
 			  ON CONFLICT (message_id, author_id) 
 			  DO UPDATE SET emoji_name = :emoji_name 
 			  WHERE v.message_id = :message_id AND v.author_id = :author_id`,
		map[string]interface{}{
			"message_id": vote.MessageID,
			"emoji_name": castVote.EmojiName,
			"author_id":  castVote.AuthorID,
		},
	)

	if dbErr != nil {
		return dbErr
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return NoRowsAffected
	}
	return nil
}
