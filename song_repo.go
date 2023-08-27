package main

import (
	"context"
	"database/sql"
)

type SongRepo struct {
	conn *sql.DB
}

func NewRepo() (repo SongRepo) {
	return
}

func (c *SongRepo) Close() {
	c.conn.Close()
}

func (s *SongRepo) Open(driverName, dataSourceName string) (err error) {
	db, err := sql.Open("sqlite3", "data.sqlite")
	if err != nil {
		return
	}
	s.conn = db
	return
}

func (s *SongRepo) FindByHash(ctx context.Context, hash string) (tx *sql.Tx, row *sql.Row, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	row = tx.QueryRowContext(ctx, "SELECT * FROM songs WHERE hash = ?", hash)
	return
}

func (s *SongRepo) FindAlike(ctx context.Context, query string) (tx *sql.Tx, rows *sql.Rows, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	queryString := `SELECT * FROM songs
                        WHERE LOWER(title) LIKE '%' || LOWER(?) || '%' 
                           OR LOWER(artist) LIKE '%' || LOWER(?) || '%' 
                           OR hash = ?`
	rows, err = tx.QueryContext(ctx, queryString, query, query, query)
	if err != nil {
		return
	}

	return
}

func (s *SongRepo) FindById(ctx context.Context, id int64) (tx *sql.Tx, row *sql.Row, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	row = tx.QueryRowContext(ctx, "SELECT * FROM songs WHERE id = ?", id)
	return
}

func (s *SongRepo) All(ctx context.Context) (tx *sql.Tx, rows *sql.Rows, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	rows, err = tx.QueryContext(ctx, "SELECT * FROM songs")
	if err != nil {
		return
	}

	return
}

func (s *SongRepo) Update(ctx context.Context, song SongData) (tx *sql.Tx, res sql.Result, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	stmt, err := tx.PrepareContext(ctx, "UPDATE songs SET title = ?, artist = ?, hash = ?, deleted = ? WHERE id = ?")
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err = tx.StmtContext(ctx, stmt).Exec(song.Title, song.Artist, song.Hash, song.Deleted, song.Id)
	if err != nil {
		return
	}

	return
}

func (s *SongRepo) Store(ctx context.Context, song SongData) (tx *sql.Tx, res sql.Result, err error) {
	tx, err = s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO songs (title, artist, hash, deleted) VALUES (?, ?, ?, FALSE)")
	if err != nil {
		return
	}
	defer stmt.Close()

	res, err = tx.StmtContext(ctx, stmt).Exec(song.Title, song.Artist, song.Hash)
	if err != nil {
		return
	}

	return
}
