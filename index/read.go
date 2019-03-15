package index

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/awbraunstein/setlist-search/searcher"
	"github.com/pkg/errors"
)

// The index format is as follows:
//
//  "setsearcher index 1"
//  [SONGS]
//  [SETLISTS]
//
// Each section will be separated by a newline ("\n") and start with the
// section's name and end with the closing marker [END]
//
// Songs will be a map from human name to short name. This will account for when
// there are multiple versions of the songs human name. The separator is "|".
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
	songs map[string]string
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
		songs:        make(map[string]string),
		setlists:     make(map[string]*searcher.Setlist),
		reverseIndex: make(map[string][]string),
	}

	scanner := bufio.NewScanner(file)
	if err := readHeader(scanner); err != nil {
		return nil, err
	}
	if err := i.readSongs(scanner); err != nil {
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

func (i *Index) readSongs(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		// We are in the songs stanza.
		if scanner.Text() == "[SONGS]" {
			continue
		}
		// We are now done with this section.
		if scanner.Text() == "[END]" {
			return nil
		}
		// Expect a longText,data-value format.
		parts := strings.Split(scanner.Text(), "|")

		if len(parts) != 2 {
			return fmt.Errorf("Song malformatted: %#v", scanner.Text())
		}
		i.songs[parts[0]] = parts[1]
	}
	return errors.New("Expected a closing statement for the songs section")
}

func (i *Index) readSetlists(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		// We are in the setlist stanza.
		if scanner.Text() == "[SETLISTS]" {
			continue
		}
		// We are now done with this section.
		if scanner.Text() == "[END]" {
			return nil
		}
		setlist := scanner.Text()
		sl, err := searcher.ParseSetlist(setlist)
		if err != nil {
			return err
		}
		i.setlists[sl.ShowId] = sl
		for _, song := range sl.Songs() {
			i.reverseIndex[song] = append(i.reverseIndex[song], sl.ShowId)
		}
	}
	return errors.New("Expected a closing statement for the setlists section")
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
