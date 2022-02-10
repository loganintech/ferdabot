CREATE TABLE ferda
(
    id        SERIAL PRIMARY KEY NOT NULL UNIQUE,
    userid    BIGINT             NOT NULL,
    creatorid BIGINT             NOT NULL,
    time      TIMESTAMPTZ        NOT NULL,
    reason    TEXT               NOT NULL
);

CREATE INDEX userid_idx ON ferda (userid);
CREATE INDEX creatorid_idx ON ferda (creatorid);

CREATE TABLE config
(
    key TEXT PRIMARY KEY NOT NULL UNIQUE,
    val TEXT             NOT NULL
);

CREATE TABLE reminder
(
    id        SERIAL PRIMARY KEY NOT NULL UNIQUE,
    creatorid TEXT               NOT NULL,
    time      TIMESTAMPTZ        NOT NULL,
    message   TEXT               NOT NULL
);
