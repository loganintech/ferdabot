CREATE TABLE ferda
(
    id        SERIAL PRIMARY KEY NOT NULL UNIQUE,
    userid    BIGINT             NOT NULL,
    creatorid BIGINT             NOT NULL,
    time      TIMESTAMPTZ        NOT NULL,
    reason    TEXT               NOT NULL
);

CREATE INDEX userid_idx ON ferda (userid);
