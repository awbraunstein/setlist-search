package searcher

import (
	"fmt"
	"regexp"
	"strings"
)

// Setlist holds the set from a show.
type Setlist struct {
	showId string
	sets   []Set
	encore Set
}

// Set holds a single set of a setlist.
type Set struct {
	songs []string
}

var (
	idRe     = regexp.MustCompile(`^ID\{(.+?)\}`)
	setRe    = regexp.MustCompile(`(?:SET\d+\{(.+?)\})+`)
	encoreRe = regexp.MustCompile(`(?:ENCORE\{(.+?)\})`)
)

// ParseSetlist parses setlists that are of the form:
//  ID{showid}SET1{song1,song2,song3,song4}SET2{songa,songb,songc,songd}ENCORE{songx,songy,songz}
func ParseSetlist(setlist string) (*Setlist, error) {
	idMatches := idRe.FindStringSubmatch(setlist)
	if len(idMatches) != 2 {
		return nil, fmt.Errorf("ParseSetlist: couldn't find ID tag at start of setlist: %s", setlist)
	}
	sl := &Setlist{
		showId: idMatches[1],
	}
	setMatches := setRe.FindStringSubmatch(setlist)
	if len(setMatches) == 0 {
		return nil, fmt.Errorf("ParseSetList: couldn't find any sets in the setlist: %s", setlist)
	}
	for i := 1; i < len(setMatches); i++ {
		s := Set{
			songs: strings.Split(setMatches[i], ","),
		}
		sl.sets = append(sl.sets, s)
	}
	encoreMatches := encoreRe.FindStringSubmatch(setlist)
	if len(encoreMatches) == 2 {
		sl.encore.songs = strings.Split(encoreMatches[1], ",")
	}
	return sl, nil
}
