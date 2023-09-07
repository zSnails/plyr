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
		}).Info("Something in the way")
		h.ServeHTTP(w, r)
	})
}

func deletedMW(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		hash := vars["hash"]
		log := logrus.WithContext(r.Context()).WithField("hash", hash)

		song := cache.Get(hash)
		if song != nil && song.Deleted {
			log.Warnf("Trying to access a deleted song.")
			w.WriteHeader(http.StatusGone)
			return
		}

		if song == nil {
			log.Warnf("Trying to access an unknown song")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	}
}
