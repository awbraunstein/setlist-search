package index

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/awbraunstein/setlist-search/searcher"
)

type IndexWriter struct {
	file *os.File
	// shows is a map from showId to show date.
	shows map[int]string
	// songs is the set of songs played in the set of shows with setlists.
	songs map[string]bool
	// setlists is a map from showid to setlist info.
	setlists map[string]*searcher.Setlist
}

func NewWriter(file *os.File) *IndexWriter {
	return &IndexWriter{
		file:     file,
		shows:    make(map[int]string),
		songs:    make(map[string]bool),
		setlists: make(map[string]*searcher.Setlist),
	}
}

func (w *IndexWriter) AddShow(id int, date string) {
	w.shows[id] = date
}

func (w *IndexWriter) AddSetlist(sl *searcher.Setlist) {
	w.setlists[sl.ShowId] = sl
	for _, set := range sl.Sets {
		for _, song := range set.Songs {
			w.songs[song] = true
		}
	}
	if sl.Encore != nil {
		for _, song := range sl.Encore.Songs {
			w.songs[song] = true
		}
	}
}

func (w *IndexWriter) writeNull() error {
	_, err := w.file.Write([]byte("\x00"))
	return err
}

func (w *IndexWriter) Write() error {
	// Reset the file.
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	if _, err := w.file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := w.file.WriteString("setsearcher index 1\n"); err != nil {
		return err
	}
	if err := w.writeNull(); err != nil {
		return err
	}

	var showIds []int
	for key := range w.shows {
		showIds = append(showIds, key)
	}

	sort.Ints(showIds)
	for _, id := range showIds {
		if _, err := fmt.Fprintf(w.file, "%d,s\n", id, w.shows[id]); err != nil {
			return err
		}
	}
	if err := w.writeNull(); err != nil {
		return err
	}

	var songs []string
	for song := range w.songs {
		songs = append(songs, song)
	}
	sort.Strings(songs)
	if _, err := w.file.WriteString(strings.Join(songs, "\n")); err != nil {
		return err
	}
	if err := w.writeNull(); err != nil {
		return err
	}
	var setlists []string
	for _, id := range showIds {
		setlists = append(setlists, w.setlists[strconv.Itoa(id)].String())
	}
	if _, err := w.file.WriteString(strings.Join(setlists, "\n")); err != nil {
		return err
	}

	return nil
}
