package storage

import (
	"fmt"

	"github.com/fatih/color"
)

type SongData struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Hash     string `json:"hash"`
	Duration int64  `json:"duration"`
	Genre    string `json:"genre"`
	Deleted  bool   `json:"-"`
}

var (
	c = color.New(color.FgWhite, color.CrossedOut)
)

func (s SongData) String() string {
	data := fmt.Sprintf("(%d) [%s]: %s - %s", s.Id, s.Hash, s.Artist, s.Title)
	if s.Deleted {
		data = c.Sprintf(data)
	}
	return data
}

type scanner interface {
	Scan(...any) error
}

func (s *SongData) FromRow(row scanner) error {
	return row.Scan(&s.Id, &s.Title, &s.Artist, &s.Hash, &s.Duration, &s.Genre, &s.Deleted)
}
