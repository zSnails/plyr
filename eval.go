package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/google/uuid"
	"github.com/hajimehoshi/go-mp3"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sirupsen/logrus"
	"github.com/zSnails/plyr/storage"
)

func eval(ctx context.Context, commandLine string, reader *bufio.Reader) (err error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logrus.WithContext(ctx).WithField("command", commandLine)

	switch commandLine {
	case "add":

		fmt.Print("File> ")
		filename, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		filename = strings.TrimSuffix(filename, "\n")

		if path.Ext(filename) != ".mp3" {
			return errors.New("Only mp3 files are supported.")
		}

		log.WithField("filename", filename).Info("Opening file.")

		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		meta, err := tag.ReadFrom(file)
		if err != nil {
			return err
		}

		// XXX: getting song duration from file
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			return err
		}

		samples := decoder.Length() / 4
		duration := samples / int64(decoder.SampleRate())

		songname := meta.Title()
		if songname == "" {
			log.Infoln("Could not read song name from file.")
			fmt.Print("Song Name> ")
			songname, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			songname = strings.TrimSuffix(songname, "\n")
		}

		artist := meta.Artist()
		if artist == "" {
			log.Infoln("Could not read song artist from file.")
			fmt.Print("Artist> ")
			artist, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			artist = strings.TrimSuffix(artist, "\n")
		}

		genre := meta.Genre()
		if genre == "" {
			log.Infoln("Could not read song genre from file.")
			fmt.Print("Genre> ")
			genre, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			genre = strings.TrimSuffix(genre, "\n")
		}

		fmt.Printf("songLength: %v\n", duration)

		id := uuid.NewMD5(uuid.NameSpaceURL, []byte(songname+artist))
		log.Infof("Assigned uuid(%s) to song\n", id)

		p := path.Join(songsDirectory, id.String())

		command := buildFfmpegCommand(p, filename)

		song := storage.SongData{
			Title:    songname,
			Artist:   artist,
			Hash:     id.String(),
			Duration: duration,
			Genre:    genre,
			Deleted:  false,
		}

		tx, res, err := repo.Store(ctx, song)
		if err != nil {
			return err
		}
		defer tx.Commit()

		if rows, _ := res.RowsAffected(); rows > 0 {
			log.WithField("file", filename).Info("Generating HLS data...")
			err = os.MkdirAll(p, os.ModePerm)
			if err != nil {
				return err
			}
			log.WithField("command", fmt.Sprintf("ffmpeg %s", command)).Info("Running Command.")
			cmd := exec.CommandContext(ctx, "ffmpeg", command...)
			err = cmd.Run()
			if err != nil {
				log.Info("An error occurred, cancelling...")
				return err
			}
			cache.StoreIfNotExists(song.Hash, &song)
		}

		log.Info("Committing transaction...")
		log.Info("Done!")

	case "delete": // WARNING: hard deletions
		log.Warn("This action is irreversible. Be careful!")
		err := eval(ctx, "all", reader)
		if err != nil {
			return err
		}
		fmt.Print("Id> ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		id = strings.TrimSuffix(id, "\n")
		uid, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		songs := cache.Filter(func(v *storage.SongData) bool {
			return v.Id == int64(uid)
		})

		var found *storage.SongData
		if len(songs) > 0 {
			found = songs[0]
		}

		tx, res, err := repo.Delete(ctx, *found)
		if err != nil {
			return err
		}

		defer tx.Commit()

		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if affected > 0 {
			log.Info("Song deleted from database. Deleting local files.")
			cache.Delete(found.Hash)
			localPath := path.Join(songsDirectory, found.Hash)
			log.WithField("local-path", localPath).Println()
			err = os.RemoveAll(localPath)
			if err != nil {
				return err
			}
		} else {
			log.Info("Nothing to delete.")
		}

	case "toggle": // deletions are soft, no need for hard deletions
		err := eval(ctx, "all", reader)
		if err != nil {
			return err
		}

		fmt.Print("Id> ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		id = strings.TrimSuffix(id, "\n")
		uid, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		songs := cache.Filter(func(v *storage.SongData) bool {
			return v.Id == int64(uid)
		})

		var found *storage.SongData
		if len(songs) < 1 {
			return nil
		}

		found = songs[0]

		found.Deleted = !found.Deleted
		tx, res, err := repo.Update(ctx, *found)
		if err != nil {
			return err
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		log.WithField("affected-rows", affected).Info("Done.")

	case "all":
		for _, song := range cache.All() {
			fmt.Println(song)
		}
	case "find":
		fmt.Print("Query> ")

		query, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		query = strings.TrimSuffix(query, "\n")

		songs := cache.Filter(func(sd *storage.SongData) bool {
			return fuzzy.Match(removeSpecialCharacters(query), removeSpecialCharacters(sd.Title+sd.Artist)) ||
				sd.Hash == query
		})

		for _, song := range songs {
			fmt.Println(song)
		}

	}

	return
}
