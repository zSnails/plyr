package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func deletedMW(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Check if the song has been deleted, if so then don't allow access to it
		vars := mux.Vars(r)

		tx, row, err := repo.FindByHash(r.Context(), vars["hash"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer tx.Commit()

		var song SongData
		// err = row.Scan(&song.Id, &song.Title, &song.Artist, &song.Hash, &song.Deleted)
		err = song.FromRow(row)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if song.Deleted {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		h.ServeHTTP(w, r)
	}
}
