package index

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/awbraunstein/setlist-search/searcher"
	"github.com/pkg/errors"
)

// The index format is as follows:
//
//  "setsearcher index 1"
//  setlists
//
// Each section will be separated by a newline ("\n).
//
// Setlists will be a list of setlists where each setlist is formatted according
// to the setlist serialization method separated by newlines.

const (
	header = "setsearcher index 1"
)

// Index is the setlist searcher index that's loaded into memory to run analysis
// on setlists.
type Index struct {
	// songs is the list of songs that are in the index.
	songs []string
	// setlists is a map from showid to setlist
	setlists map[string]*searcher.Setlist
	// map from song to the list of showids that that song was played in.
	reverseIndex map[string][]string
}

// Open reads an Index from disk.
func Open(filename string) (*Index, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	i := &Index{
		setlists:     make(map[string]*searcher.Setlist),
		reverseIndex: make(map[string][]string),
	}

	scanner := bufio.NewScanner(file)
	if err := readHeader(scanner); err != nil {
		return nil, err
	}
	if err := i.readSetlists(scanner); err != nil {
		return nil, err
	}
	return i, nil
}

func readHeader(scanner *bufio.Scanner) error {
	if !scanner.Scan() {
		errors.New("index contains no header")
	}
	if scanner.Text() != header {
		fmt.Errorf("index header malformed; %q", scanner.Text())
	}
	return nil
}

func (i *Index) readSetlists(scanner *bufio.Scanner) error {
	songSet := make(map[string]bool)
	for scanner.Scan() {
		setlist := scanner.Text()
		sl, err := searcher.ParseSetlist(setlist)
		if err != nil {
			return err
		}
		i.setlists[sl.ShowId] = sl
		for _, song := range sl.Songs() {
			i.reverseIndex[song] = append(i.reverseIndex[song], sl.ShowId)
			songSet[song] = true
		}
	}
	for song := range songSet {
		i.songs = append(i.songs, song)
	}
	sort.Strings(i.songs)
	if err := scanner.Err(); err != io.EOF {
		return err
	}
	return nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
