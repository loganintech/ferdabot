
CREATE TABLE IF NOT EXISTS vote
(
    message_id  TEXT PRIMARY KEY NOT NULL UNIQUE,
    title       TEXT             NOT NULL,
    description TEXT             NOT NULL,
    creator_id  TEXT             NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL,
    active      BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS vote_lines
(
    message_id TEXT NOT NULL,
    emoji_name TEXT NOT NULL,
    emoji_id   TEXT NOT NULL,
    line_value TEXT NOT NULL,
    PRIMARY KEY (message_id, emoji_name)
);

CREATE TABLE IF NOT EXISTS vote_cast
(
    message_id TEXT NOT NULL,
    emoji_name TEXT NOT NULL,
    author_id  TEXT NOT NULL,
    PRIMARY KEY (message_id, author_id)
);