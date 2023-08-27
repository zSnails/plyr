PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE songs (id integer primary key, title text, artist text, hash text unique, deleted bool);
COMMIT;
