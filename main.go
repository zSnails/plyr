package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var repo SongRepo

func init() {
	repo = NewRepo()
	if err := repo.Open("sqlite3", "data.sqlite?cache=shared&mode=rwc"); err != nil {
		panic(err)
	}
}

func filter[T any](s []T, fn func(T) bool) (result []T) {
	result = []T{}
	for _, elem := range s {
		if fn(elem) {
			result = append(result, elem)
		}
	}
	return
}

func main() {
	// configure the songs directory name and port
	const songsDir = "songs"
	const port = 8080
	defer repo.Close()

	r := mux.NewRouter()
	// add a handler for the song files

	// Music file server
	r.HandleFunc("/song/{songName}", songHandler)
	r.HandleFunc("/song/all", allSongs)
	r.Handle("/{hash}/{file}", deletedMW(http.FileServer(http.Dir(songsDir))))

	// serve and log errors
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
	}()
	inputReader := bufio.NewReader(os.Stdin)
	for { // Server menu
		fmt.Print("> ")
		line, err := inputReader.ReadString('\n')
		if err == io.EOF {
			fmt.Println()
			break
		} else if err != nil {
			panic(err)
		}

		r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
		commandLine := r.FindAllString(line, -1)
		if len(commandLine) == 0 {
			continue // Skip empty line
		}

		// TODO: implement an actual command line, the original idea was to
		// parse the command line and extract its arguments, however I got
		// carried away and didn't actually do that
		err = eval(commandLine, inputReader)
		if err != nil {
			panic(err)
		}
	}
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
