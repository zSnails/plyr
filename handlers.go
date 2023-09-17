package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/lithammer/fuzzysearch/fuzzy"
	log "github.com/sirupsen/logrus"
	zslices "github.com/zSnails/plyr/slices"
	"github.com/zSnails/plyr/storage"
)

func allSongs(w http.ResponseWriter, r *http.Request) {
	log := log.WithContext(r.Context()).WithField("handler", "all-songs")

	songs := cache.All()

	songs = zslices.Filter(songs, func(sd *storage.SongData) bool {
		return !sd.Deleted
	})

	data, err := json.Marshal(songs)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("Sending song data")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}

func songQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query, err := url.QueryUnescape(vars["songName"])
	log := log.WithContext(r.Context()).WithField("song-name", query).WithField("handler", "song-query")
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	songs := cache.Filter(func(sd *storage.SongData) bool {
		return fuzzy.Match(removeSpecialCharacters(query), removeSpecialCharacters(sd.Title+sd.Artist)) ||
			sd.Hash == query
	})

	data, err := json.Marshal(songs)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debug("Sending song data")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data)
}
