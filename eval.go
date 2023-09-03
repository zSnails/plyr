package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func eval(ctx context.Context, commandLine []string, reader *bufio.Reader) (err error) {

	var ffmpegCommand = []string{
		"-i",
		"", // 1
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
		"outputlist.m3u8",
		"-segment_format",
		"mpegts",
		"", // 16
		//"output%03d.ts",
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logrus.WithContext(ctx)

	switch commandLine[0] {
	case "add":

		fmt.Print("Song Name> ")
		songname, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		songname = strings.TrimSuffix(songname, "\n")

		fmt.Print("Artist> ")
		artist, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		artist = strings.TrimSuffix(artist, "\n")

		fmt.Print("File> ")
		file, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		id := uuid.NewMD5(uuid.NameSpaceURL, []byte(songname+artist))
		file = strings.TrimSuffix(file, "\n")

		p := path.Join("songs", id.String())

		ffmpegCommand[1] = file
		ffmpegCommand[13] = path.Join(p, ffmpegCommand[13])
		ffmpegCommand[16] = path.Join(p, "output%03d.ts")

		song := SongData{
			Title:   songname,
			Artist:  artist,
			Hash:    id.String(),
			Deleted: false,
		}

		tx, res, err := repo.Store(ctx, song)
		if err != nil {
			return err
		}
		defer tx.Commit()

		if rows, _ := res.RowsAffected(); rows > 0 {
			log.WithField("file", file).Info("Generating HLS data...")
			err = os.MkdirAll(p, os.ModePerm)
			if err != nil {
				return err
			}
			log.WithField("command", ffmpegCommand).Info("Running Command.")
			cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegCommand...)
			err = cmd.Run()
			if err != nil {
				log.Info("An error occurred, cancelling...")
				return err
			}
		}

		log.Info("Committing transaction...")
		log.Info("Done!")

	case "delete": // WARNING: hard deletions
		err := eval(ctx, []string{"all"}, reader)
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

		tx, row, err := repo.FindById(ctx, int64(uid))
		if err != nil {
			return err
		}
		var found SongData

		err = found.FromRow(row)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		tx, res, err := repo.Delete(ctx, found)
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
			err = os.RemoveAll(path.Join("songs", found.Hash))
			if err != nil {
				return err
			}
		} else {
			log.Info("Nothing to delete.")
		}

	case "toggle": // deletions are soft, no need for hard deletions
		err := eval(ctx, []string{"all"}, reader)
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

		tx, row, err := repo.FindById(ctx, int64(uid))
		if err != nil {
			return err
		}
		var found SongData

		err = found.FromRow(row)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		found.Deleted = !found.Deleted
		tx, res, err := repo.Update(ctx, found)
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
		tx, rows, err := repo.All(ctx)
		if err != nil {
			return err
		}
		defer tx.Commit()

		err = rows.Err()
		if err != nil {
			return err
		}

		songs, err := makeSongDataSlice(rows)
		if err != nil {
			return err
		}

		for _, song := range songs {
			fmt.Println(song)
		}

	case "find":
		fmt.Print("Query> ")

		query, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		query = strings.TrimSuffix(query, "\n")

		tx, rows, err := repo.FindAlike(ctx, query)
		if err != nil {
			return err
		}

		defer tx.Commit()

		err = rows.Err()
		if err != nil {
			return err
		}

		songs, err := makeSongDataSlice(rows)
		if err != nil {
			return err
		}

		for _, song := range songs {
			fmt.Println(song)
		}

	}

	return
}
