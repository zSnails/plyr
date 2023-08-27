package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func allSongs(w http.ResponseWriter, r *http.Request) {
	tx, rows, err := repo.All(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Commit()
	songs, err := makeSongDataSlice(rows)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	songs = filter(songs, func(sd SongData) bool {
		return !sd.Deleted
	})
	data, err := json.Marshal(songs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}

func songHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songName, err := url.QueryUnescape(vars["songName"])
	if err != nil {
		// fmt.Fprint(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx, rows, err := repo.FindAlike(r.Context(), songName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Commit()
	songs, err := makeSongDataSlice(rows)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	songs = filter(songs, func(sd SongData) bool {
		return !sd.Deleted
	})
	data, err := json.Marshal(songs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}
