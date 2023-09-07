package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/dhowden/tag"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/hajimehoshi/go-mp3"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/zSnails/plyr/storage"
)

type App struct {
	ui *tview.Application
	*State
}

func (a *App) OnInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case rune(tcell.KeyCtrlQ):
		a.ui.Stop()
	case rune(tcell.KeyCtrlA):
		a.SongsForm.Clear(true)

		var filename string

		a.SongsForm.AddInputField("Song File", "", 10, nil, func(text string) {
			filename = text
		})

		a.SongsForm.AddButton("Save", func() {
			if path.Ext(filename) != ".mp3" {
				logrus.Error("Only mp3 files are supported.")
			} else {
				a.AddSongModal(filename)
			}
		})

		a.SongsForm.AddButton("Cancel", func() {
			a.Pages.SwitchToPage("Songs")
		})
		a.Pages.SwitchToPage("Songs Form")
		a.SongsForm.SetFocus(0)
	}
	return event
}

func (a *App) AddSongModal(filename string) {

	a.SongsForm.Clear(true)
	log := logrus.WithField("file", filename)

	log.Debug("Opening")
	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	log.Debug("Reading tags")
	meta, err := tag.ReadFrom(file)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug("Decoding")
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		log.Error(err)
		return
	}

	samples := decoder.Length() / 4
	duration := samples / int64(decoder.SampleRate())

	song := storage.SongData{
		Duration: duration,
		Deleted:  false,
	}

	song.Title = meta.Title()

	a.SongsForm.AddInputField("Name", song.Title, 10, nil, func(text string) {
		song.Title = text
		log.Debug(song.Title)
	})

	song.Artist = meta.Artist()

	a.SongsForm.AddInputField("Artist", song.Artist, 10, nil, func(text string) {
		song.Artist = text
		log.Debug(song.Artist)
	})

	song.Genre = meta.Genre()

	a.SongsForm.AddInputField("Genre", song.Genre, 10, nil, func(text string) {
		song.Genre = text
		log.Debug(song.Genre)
	})

	log.WithField("duration", duration).Debug()

	a.SongsForm.AddButton("Save", func() {
		a.Pages.SwitchToPage("Songs")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer logrus.Debug("Canceling local context")
		song.Hash = uuid.NewMD5(uuid.NameSpaceURL, []byte(song.Title+song.Artist)).String()
		log.Infof("Assigned uuid(%s) to song\n", song.Hash)

		p := path.Join(songsDirectory, song.Hash)

		command := buildFfmpegCommand(p, filename)

		tx, res, err := repo.Store(ctx, song)
		if err != nil {
			log.Error(err)
			return
		}
		defer log.Info("Committing transaction...")
		defer tx.Commit()

		if rows, _ := res.RowsAffected(); rows > 0 {
			song.Id, err = res.LastInsertId()
			if err != nil {
				log.Error(err)
				return
			}

			log.WithField("file", filename).Info("Generating HLS data...")
			err = os.MkdirAll(p, os.ModePerm)
			if err != nil {
				log.Error(err)
				return
			}

			log.WithField("command", fmt.Sprintf("ffmpeg %s", command)).Info("Running Command.")
			cmd := exec.CommandContext(ctx, "ffmpeg", command...)
			err = cmd.Run()
			if err != nil {
				log.Error(err)
				return
			}
			log.Info("Created local processed files.")
			log.Debug("Adding new song to songs cache.")
			a.AddSong(song)
			a.Cache.StoreIfNotExists(song.Hash, &song)
		}

		log.Info("Done!")
	})
}

func (s *App) FilterSongs(query string) {
	results := s.State.Cache.Filter(func(sd *storage.SongData) bool {
		return fuzzy.Match(removeSpecialCharacters(query), removeSpecialCharacters(sd.Title+sd.Artist)) ||
			sd.Hash == query
	})

	s.SongsList.Clear()
	for _, song := range results {
		s.SongsAppend(*song)
	}
}

func (s *App) DeleteSongModal(song storage.SongData) {
	modal := tview.NewModal()
	modal.SetTitle("Importante!")
	modal.SetText("Seguro que quieres borrar esta canción?")
	modal.AddButtons([]string{"Si", "No"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			s.Pages.SwitchToPage("Songs")
			s.Pages.RemovePage("modal")
			return
		}
		s.Repo.Delete(context.Background(), song)
		s.Cache.Delete(song.Hash)
	})
	s.Pages.AddPage("modal", modal, true, true)
	s.Pages.SwitchToPage("modal")
}

func newApp(cache *storage.Cache[string, storage.SongData]) *App {

	app := &App{
		ui: tview.NewApplication(),
		State: &State{
			Pages:        tview.NewPages(),
			Cache:        cache,
			SongsFlexBox: tview.NewFlex(),
			SongsList:    tview.NewList(),
			SongsForm:    tview.NewForm(),
			Flex:         tview.NewFlex(),
			Logs:         tview.NewTextView(),
			Repo:         &storage.SongRepo{},
		},
	}

	app.ui.EnableMouse(true)
	app.ui.SetInputCapture(app.OnInput)

	app.SongsList.SetBorder(true)
	app.SongsList.SetTitle("Songs")

	app.SongsFlexBox.AddItem(tview.NewForm().AddInputField("Search", "", 20, nil, func(text string) {
		app.FilterSongs(text)
	}), 0, 1, true)

	app.SongsFlexBox.SetDirection(tview.FlexRow)

	app.SongsFlexBox.AddItem(app.SongsList, 0, 12, false)

	app.SongsForm.SetBorder(true)
	app.SongsForm.SetTitle("Songs Form")
	app.SongsForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune(tcell.KeyESC):
			app.Pages.SwitchToPage("Songs")
		}
		return event
	})

	app.Logs.SetBorder(true)
	app.Logs.SetDynamicColors(true)
	app.Logs.SetTitle("Logs")
	app.Logs.SetScrollable(true)

	app.Logs.SetChangedFunc(func() {
		app.ui.Draw()
		app.Logs.ScrollToEnd()
	})

	app.Pages.AddAndSwitchToPage("Songs", app.SongsFlexBox, true)
	app.Pages.AddPage("Songs Form", app.SongsForm, true, false)

	app.Flex.AddItem(app.Pages, 0, 1, true)
	app.Flex.AddItem(app.Logs, 0, 2, false)
	app.ui.SetRoot(app.Flex, true)

	app.State.App = app

	for _, song := range app.Cache.All() {
		app.AddSong(*song)
	}

	return app
}
