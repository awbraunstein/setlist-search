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
	// setlists is a map from showid to setlist info.
	setlists map[string]*searcher.Setlist
	songs    map[string]string

	// The tmp file we will be writing to.
	file *os.File
}

func NewWriter(indexLocation string) *IndexWriter {
	return &IndexWriter{
		indexLocation: indexLocation,
		setlists:      make(map[string]*searcher.Setlist),
		songs:         make(map[string]string),
	}
}

func (w *IndexWriter) AddSetlist(sl *searcher.Setlist) {
	w.setlists[sl.ShowId] = sl
}

func (w *IndexWriter) AddSong(songName, songValue string) {
	w.songs[songName] = songValue
}

func (w *IndexWriter) Write() error {
	var err error
	w.file, err = ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	if _, err := w.file.WriteString(header); err != nil {
		return err
	}
	if _, err := w.file.WriteString("\n"); err != nil {
		return err
	}

	var songNames []string
	for name := range w.songs {
		songNames = append(songNames, name)
	}
	sort.Strings(songNames)

	var songs []string
	for _, name := range songNames {
		songs = append(songs, fmt.Sprintf("%s|%s", name, w.songs[name]))
	}
	w.file.WriteString("[SONGS]\n")
	if _, err := w.file.WriteString(strings.Join(songs, "\n")); err != nil {
		return err
	}
	if len(songs) > 0 {
		w.file.WriteString("\n")
	}
	w.file.WriteString("[END]\n")

	var showIds []string
	for key := range w.setlists {
		showIds = append(showIds, key)
	}

	sort.Strings(showIds)
	var setlists []string
	for _, id := range showIds {
		setlists = append(setlists, w.setlists[id].String())
	}
	w.file.WriteString("[SETLISTS]\n")
	if _, err := w.file.WriteString(strings.Join(setlists, "\n")); err != nil {
		return err
	}
	if len(setlists) > 0 {
		w.file.WriteString("\n")
	}
	w.file.WriteString("[END]")
	w.file.Close()
	return os.Rename(w.file.Name(), w.indexLocation)
}

func (i *Index) Write(indexLocation string) error {
	iw := NewWriter(indexLocation)
	for _, setlist := range i.setlists {
		iw.AddSetlist(setlist)
	}
	for songName, songValue := range i.songs {
		iw.AddSong(songName, songValue)
	}
	return iw.Write()
}
