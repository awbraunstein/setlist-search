package index

import (
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

	// The tmp file we will be writing to.
	file *os.File
}

func NewWriter(indexLocation string) *IndexWriter {
	return &IndexWriter{
		indexLocation: indexLocation,
		setlists:      make(map[string]*searcher.Setlist),
	}
}

func (w *IndexWriter) AddSetlist(sl *searcher.Setlist) {
	w.setlists[sl.ShowId] = sl
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
	if _, err := w.file.WriteString(header); err != nil {
		return err
	}
	if _, err := w.file.WriteString("\n"); err != nil {
		return err
	}

	var showIds []string
	for key := range w.setlists {
		showIds = append(showIds, key)
	}

	sort.Strings(showIds)
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

func (i *Index) Write(indexLocation string) error {
	iw := NewWriter(indexLocation)
	for _, setlist := range i.setlists {
		iw.AddSetlist(setlist)
	}
	return iw.Write()
}
