PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE songs (
    id integer primary key,
    title text,
    artist text,
    hash text unique,
    duration integer,
    genre text,
    deleted bool
);
COMMIT;
