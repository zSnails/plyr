package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var (
	reg            = regexp.MustCompile(`[^a-zA-Z0-9áéíóúÁÉÍÓÚ]`)
	repo           SongRepo
	songsDirectory string
	port           string
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

	repo = NewRepo()

	app = newApp()

	app.SetSongRepo(&repo)
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

func filter[T any](s []T, fn func(T) bool) []T {
	result := []T{}
	for _, elem := range s {
		if fn(elem) {
			result = append(result, elem)
		}
	}
	return result
}

func main() {
	flag.Parse()
	defer repo.Close()

	r := mux.NewRouter()

	s := r.PathPrefix("/song").Subrouter()
	{
		s.HandleFunc("", allSongs)
		s.HandleFunc("/{songName}", songHandler)
	}
	s.Use(loggerMW)
	r.Handle("/{hash}/{file}", deletedMW(http.FileServer(http.Dir(songsDirectory))))

	ctx := context.Background()

	log := logrus.WithContext(ctx)
	go func() {
		log.Fatal(http.ListenAndServe(":"+port, r))
	}()

	err := app.CacheSongs()
	if err != nil {
		log.Panic(err)
	}

	logrus.SetOutput(app.Logs) // set the output to the logs window
	if err := app.ui.Run(); err != nil {
		log.Panic(err)
	}

	// inputReader := bufio.NewReader(os.Stdin)
	// for { // Server menu
	// 	fmt.Print(">>> ")
	// 	line, err := inputReader.ReadString('\n')
	// 	if err == io.EOF {
	// 		fmt.Println()
	// 		break
	// 	} else if err != nil {
	// 		logrus.Panic(err)
	// 	}

	// 	r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	// 	commandLine := r.FindAllString(line, -1)
	// 	if len(commandLine) == 0 {
	// 		continue // Skip empty line
	// 	}

	// 	// TODO: implement an actual command line, the original idea was to
	// 	// parse the command line and extract its arguments, however I got
	// 	// carried away and didn't actually do that
	// 	err = eval(ctx, commandLine, inputReader)
	// 	if err != nil {
	// 		log.Error(err)
	// 	}
	// }

}

// NOTE: future self, the reason I'm not using a named parameter here is
// because the marshaler will default to null and I don't want that
func makeSongDataSlice(rows *sql.Rows) ([]SongData, error) {
	result := []SongData{}
	for rows.Next() {
		var songData SongData
		err := songData.FromRow(rows)
		if err != nil {
			return result, err
		}
		result = append(result, songData)
	}
	return result, nil
}
