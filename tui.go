package main

import (
	"context"
	"path"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type App struct {
	ui *tview.Application
	*State
}

type State struct {
	cachedsongs []SongData
	pages       *tview.Pages
	songsPage   *tview.List
	songsForm   *tview.Form
	flex        *tview.Flex
	logs        *tview.TextView
	repo        *SongRepo
}

func (a *App) OnInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case rune(tcell.KeyCtrlQ):
		a.ui.Stop()
	case rune(tcell.KeyCtrlA):
		current := a.pages.GetTitle()
		logrus.Debug(current)
		if current == "Songs Form" {
			logrus.Warn("Already there")
			return event
		}
		logrus.Debug("Pressed Ctrl+a")

		var filename string
		a.songsForm.AddInputField("Song File", "", 10, nil, func(text string) {
			filename = text
			logrus.Debug(filename)
		})
		a.songsForm.AddButton("accept", func() {
			if path.Ext(filename) != ".mp3" {
				logrus.Error("Only mp3 files are supported.")
			} else {
				a.songsForm.Clear(true)
				a.AddSong(filename)
			}
		})
		a.pages.SwitchToPage("Songs Form")
	}
	return event
}

func (a *App) AddSong(filename string) {
	// fmt.Print("File> ")

	// log.WithField("filename", filename).Info("Opening file.")

	// file, err := os.Open(filename)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// meta, err := tag.ReadFrom(file)
	// if err != nil {
	// 	return err
	// }

	// // XXX: getting song duration from file
	// decoder, err := mp3.NewDecoder(file)
	// if err != nil {
	// 	return err
	// }

	// samples := decoder.Length() / 4
	// duration := samples / int64(decoder.SampleRate())

	// songname := meta.Title()
	// if songname == "" {
	// 	log.Infoln("Could not read song name from file.")
	// 	fmt.Print("Song Name> ")
	// 	songname, err = reader.ReadString('\n')
	// 	if err != nil {
	// 		return err
	// 	}
	// 	songname = strings.TrimSuffix(songname, "\n")
	// }

	// artist := meta.Artist()
	// if artist == "" {
	// 	log.Infoln("Could not read song artist from file.")
	// 	fmt.Print("Artist> ")
	// 	artist, err = reader.ReadString('\n')
	// 	if err != nil {
	// 		return err
	// 	}
	// 	artist = strings.TrimSuffix(artist, "\n")
	// }

	// genre := meta.Genre()
	// if genre == "" {
	// 	log.Infoln("Could not read song genre from file.")
	// 	fmt.Print("Genre> ")
	// 	genre, err = reader.ReadString('\n')
	// 	if err != nil {
	// 		return err
	// 	}
	// 	genre = strings.TrimSuffix(genre, "\n")
	// }

	// fmt.Printf("songLength: %v\n", duration)

	// id := uuid.NewMD5(uuid.NameSpaceURL, []byte(songname+artist))
	// log.Infof("Assigned uuid(%s) to song\n", id)

	// p := path.Join(songsDirectory, id.String())

	// ffmpegCommand[1] = filename
	// ffmpegCommand[13] = path.Join(p, ffmpegCommand[13])
	// ffmpegCommand[16] = path.Join(p, "output%03d.ts")

	// song := SongData{
	// 	Title:    songname,
	// 	Artist:   artist,
	// 	Hash:     id.String(),
	// 	Duration: duration,
	// 	Genre:    genre,
	// 	Deleted:  false,
	// }

	// tx, res, err := repo.Store(ctx, song)
	// if err != nil {
	// 	return err
	// }
	// defer tx.Commit()

	// if rows, _ := res.RowsAffected(); rows > 0 {
	// 	log.WithField("file", filename).Info("Generating HLS data...")
	// 	err = os.MkdirAll(p, os.ModePerm)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	log.WithField("command", fmt.Sprintf("ffmpeg %s", ffmpegCommand)).Info("Running Command.")
	// 	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegCommand...)
	// 	err = cmd.Run()
	// 	if err != nil {
	// 		log.Info("An error occurred, cancelling...")
	// 		return err
	// 	}
	// }

	// log.Info("Committing transaction...")
	// log.Info("Done!")
}

func (s *State) SetSongRepo(repo *SongRepo) {
	s.repo = repo
}

func (s *State) CacheSongs() error {
	tx, rows, err := s.repo.All(context.Background())
	if err != nil {
		return err
	}
	defer tx.Commit()
	for rows.Next() {
		var sd SongData
		sd.FromRow(rows)
		s.AddSong(sd)
	}
	return nil
}

func (s *State) AddSong(song SongData) {
	s.cachedsongs = append(s.cachedsongs, song)
	s.songsPage.AddItem(song.Title, song.Artist, rune(song.Id), func() {
		logrus.Debugf("Clicked %s\n", song)
		s.songsPage.SetTitle(song.Title)
	})
}

func newApp() *App {

	app := &App{
		ui: tview.NewApplication(),
		State: &State{
			cachedsongs: []SongData{},
			pages:       tview.NewPages(),
			songsPage:   tview.NewList(),
			songsForm:   tview.NewForm(),
			flex:        tview.NewFlex(),
			logs:        tview.NewTextView(),
			repo:        &SongRepo{},
		},
	}

	app.ui.EnableMouse(true)
	app.ui.SetInputCapture(app.OnInput)

	app.songsPage.SetBorder(true)
	app.songsPage.SetTitle("Songs")

	app.songsForm.SetBorder(true)
	app.songsForm.SetTitle("Songs Form")

	app.logs.SetBorder(true)
	app.logs.SetDynamicColors(true)
	app.logs.SetTitle("Logs")
	app.logs.SetScrollable(true)

	app.logs.SetChangedFunc(func() {
		app.ui.Draw()
		app.logs.ScrollToEnd()
	})

	app.pages.AddAndSwitchToPage("Songs", app.songsPage, true)
	app.pages.AddPage("Songs Form", app.songsForm, true, false)

	app.flex.AddItem(app.pages, 0, 1, true)
	app.flex.AddItem(app.logs, 0, 2, false)
	app.ui.SetRoot(app.flex, true)

	return app
}
