package main

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/zSnails/plyr/storage"
)

type State struct {
	App *App

	Cache *storage.Cache[string, storage.SongData]
	Repo  *storage.SongRepo

	Pages        *tview.Pages
	SongsList    *tview.List
	SongsFlexBox *tview.Flex
	SongsForm    *tview.Form
	Flex         *tview.Flex
	Logs         *tview.TextView
}

func (s *State) SetSongRepo(repo *storage.SongRepo) {
	s.Repo = repo
}

func (s *State) AddSong(song storage.SongData) {
	s.SongsAppend(song)
}

func (s *State) SongsAppend(song storage.SongData) {
	if song.Deleted {
		return
	}
	s.SongsList.AddItem(fmt.Sprintf("%s - %s", song.Artist, song.Title), song.Hash, rune(song.Title[0]), func() {
		s.App.DeleteSongModal(song)
	})
}
