package index

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/awbraunstein/setlist-search/searcher"
)

type IndexWriter struct {
	// indexLocation is the location that the index will be written to.
	indexLocation string
	// shows is a map from showId to show date.
	shows map[string]string
	// songs is the set of songs played in the set of shows with setlists.
	songs map[string]bool
	// setlists is a map from showid to setlist info.
	setlists map[string]*searcher.Setlist

	// The tmp file we will be writing to.
	file *os.File
}

func NewWriter(indexLocation string) *IndexWriter {
	return &IndexWriter{
		indexLocation: indexLocation,
		shows:         make(map[string]string),
		songs:         make(map[string]bool),
		setlists:      make(map[string]*searcher.Setlist),
	}
}

func (w *IndexWriter) AddShow(id string, date string) {
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
	var err error
	w.file, err = ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	if _, err := w.file.WriteString("setsearcher index 1\n"); err != nil {
		return err
	}
	if err := w.writeNull(); err != nil {
		return err
	}

	var showIds []string
	for key := range w.shows {
		showIds = append(showIds, key)
	}

	sort.Strings(showIds)
	for _, id := range showIds {
		if _, err := fmt.Fprintf(w.file, "%s,%s\n", id, w.shows[id]); err != nil {
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
		set, ok := w.setlists[id]
		if ok {
			setlists = append(setlists, set.String())
		}
	}
	if _, err := w.file.WriteString(strings.Join(setlists, "\n")); err != nil {
		return err
	}
	w.file.Close()
	return os.Rename(w.file.Name(), w.indexLocation)
}
