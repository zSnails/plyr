package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/zSnails/plyr/storage"
)

var (
	reg            = regexp.MustCompile(`[^a-zA-Z0-9áéíóúÁÉÍÓÚ]`)
	repo           storage.SongRepo
	cache          storage.Cache[string, storage.SongData]
	songsDirectory string
	port           string
	isTui          bool
	app            *App
)

func buildFfmpegCommand(pth, filename string) []string {
	return []string{
		"-i",
		filename,
		"-c:a",
		"libmp3lame",
		"-b:a",
		"128k",
		"-map",
		"0:0",
		"-f",
		"segment",
		"-segment_time",
		"10",
		"-segment_list",
		path.Join(pth, "outputlist.m3u8"),
		"-segment_format",
		"mpegts",
		path.Join(pth, "output%03d.ts"),
	}
}

func removeSpecialCharacters(input string) string {
	result := reg.ReplaceAllString(input, "")
	return strings.ToLower(result)
}

func init() {
	flag.StringVar(&songsDirectory, "songs-directory", "processed", "The directory where the processed songs will be stored")
	flag.StringVar(&port, "port", "8080", "The port where the server will listen")
	flag.BoolVar(&isTui, "tui", false, "Whether or not to start a tui environment")

	repo = storage.NewRepo()
	cache = storage.NewCache[string, storage.SongData]()

	logrus.SetLevel(logrus.DebugLevel)

	sql.Register("sqlite3_custom", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("remove_special_characters", removeSpecialCharacters, false); err != nil {
				return err
			}
			return nil
		},
	})

	if err := repo.Open("sqlite3_custom", "data.sqlite?cache=shared&mode=rwc"); err != nil {
		logrus.Panic(err)
	}
}

func CacheSongs(cache *storage.Cache[string, storage.SongData]) error {
	tx, rows, err := repo.All(context.Background())
	if err != nil {
		return err
	}

	defer tx.Commit()
	for rows.Next() {
		var song storage.SongData
		song.FromRow(rows)
		cache.StoreIfNotExists(song.Hash, &song)
	}
	return nil
}

func main() {
	flag.Parse()
	defer repo.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, os.Kill, os.Interrupt)

	go func() {
		<-c
		cancel()
		os.Exit(0)
	}()

	r := mux.NewRouter()

	s := r.PathPrefix("/api/song").Subrouter()
	{
		s.HandleFunc("", allSongs)
		s.HandleFunc("/{songName}", songHandler)
	}
	s.Use(loggerMW)
	r.Handle("/{hash}/{file}", deletedMW(http.FileServer(http.Dir(songsDirectory))))

	log := logrus.WithContext(ctx)
	go func() {
		log.Fatal(http.ListenAndServe(":"+port, r))
	}()

	err := CacheSongs(&cache)
	if err != nil {
		log.Fatal(err)
	}

	if isTui {
		app = newApp(&cache)
		app.SetSongRepo(&repo)

		log.Logger.SetOutput(app.Logs) // set the output to the logs window
		if err := app.ui.Run(); err != nil {
			log.Panic(err)
		}

	} else {
		inputReader := bufio.NewReader(os.Stdin)
		for { // Server menu
			fmt.Print(">>> ")
			line, err := inputReader.ReadString('\n')
			if err == io.EOF {
				fmt.Println()
				break
			} else if err != nil {
				log.Panic(err)
			}

			line = strings.Fields(strings.TrimSuffix(line, "\n"))[0]

			err = eval(ctx, line, inputReader)
			if err != nil {
				log.Error(err)
			}
		}
	}
}
