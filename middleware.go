package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func loggerMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithContext(r.Context()).WithFields(logrus.Fields{
			"remote-address": r.RemoteAddr,
			"request-uri":    r.RequestURI,
			"user-agent":     r.Header.Get("User-Agent"),
		}).Info()
		h.ServeHTTP(w, r)
	})
}

func deletedMW(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Check if the song has been deleted, if so then don't allow access to it
		vars := mux.Vars(r)
		log := logrus.WithContext(r.Context()).WithField("hash", vars["hash"])

		tx, row, err := repo.FindByHash(r.Context(), vars["hash"])
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer tx.Commit()

		var song SongData
		err = song.FromRow(row)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if song.Deleted {
			log.Warnf("Trying to access a deleted song.")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	}
}
