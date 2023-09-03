package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func allSongs(w http.ResponseWriter, r *http.Request) {
	log := log.WithContext(r.Context())
	tx, rows, err := repo.All(r.Context())
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	songs, err := makeSongDataSlice(rows)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	songs = filter(songs, func(sd SongData) bool {
		return !sd.Deleted
	})
	data, err := json.Marshal(songs)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}

func songHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songName, err := url.QueryUnescape(vars["songName"])
	log := log.WithContext(r.Context()).WithField("song-name", songName)
	if err != nil {
		// fmt.Fprint(w, err)
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tx, rows, err := repo.FindAlike(r.Context(), songName)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Commit()
	songs, err := makeSongDataSlice(rows)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	songs = filter(songs, func(sd SongData) bool {
		return !sd.Deleted
	})
	data, err := json.Marshal(songs)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}
