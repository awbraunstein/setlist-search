package searcher

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Setlist holds the set from a show.
type Setlist struct {
	ShowId string
	Date   string
	Sets   []*Set
	Encore *Set
}

// Set holds a single set of a setlist.
type Set struct {
	Songs []string
}

var (
	idRe     = regexp.MustCompile(`^ID\{(.+?)\}`)
	dateRe   = regexp.MustCompile(`DATE\{(.+?)\}`)
	setRe    = regexp.MustCompile(`(?:SET\d+\{(.*?)\})`)
	encoreRe = regexp.MustCompile(`(?:ENCORE\{(.*?)\})`)
)

// ParseSetlist parses setlists that are of the form:
//  ID{showid}SET1{song1,song2,song3,song4}SET2{songa,songb,songc,songd}ENCORE{songx,songy,songz}
func ParseSetlist(setlist string) (*Setlist, error) {
	idMatches := idRe.FindStringSubmatch(setlist)
	if len(idMatches) != 2 {
		return nil, fmt.Errorf("ParseSetlist: couldn't find ID tag at start of setlist: %s", setlist)
	}
	dateMatches := dateRe.FindStringSubmatch(setlist)
	if len(dateMatches) != 2 {
		return nil, fmt.Errorf("ParseSetlist: couldn't find DATE tag in setlist: %s", setlist)
	}
	sl := &Setlist{
		ShowId: idMatches[1],
		Date:   dateMatches[1],
	}
	setMatches := setRe.FindAllStringSubmatch(setlist, -1)
	if len(setMatches) == 0 {
		return nil, fmt.Errorf("ParseSetList: couldn't find any sets in the setlist: %s", setlist)
	}
	for _, match := range setMatches {
		s := &Set{
			Songs: strings.Split(match[1], ","),
		}
		sl.Sets = append(sl.Sets, s)
	}
	encoreMatches := encoreRe.FindStringSubmatch(setlist)
	if len(encoreMatches) == 2 {
		sl.Encore = &Set{Songs: strings.Split(encoreMatches[1], ",")}
	}
	return sl, nil
}

func ParseSetlistFromPhishNet(showId, date, setlist string) (*Setlist, error) {
	sl := &Setlist{
		ShowId: showId,
		Date:   date,
	}
	root, err := html.Parse(strings.NewReader(setlist))
	if err != nil {
		return nil, err
	}

	getSongs := func(n *html.Node, set *Set) {
		for current := n; current != nil; current = current.NextSibling {
			isSong := false
			if current.DataAtom == atom.A {
				for _, attr := range current.Attr {
					if attr.Key == "class" && attr.Val == "setlist-song" {
						isSong = true
						break
					}
				}
			}
			if isSong {
				var name string
				for _, attr := range current.Attr {
					if attr.Key == "href" {
						name = strings.TrimPrefix(attr.Val, "http://phish.net/song/")
						break
					}
				}
				set.Songs = append(set.Songs, name)
			}

		}
	}

	// Setlists are organized like:
	// <p><span class='set-label'>Set 1</span>: <a href="http://phish.net/song/nicu">NICU</a> ... </p>
	var findSets func(n *html.Node)
	findSets = func(n *html.Node) {
		foundSetInfo := false
		if n.DataAtom == atom.Span {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "set-label" {
					foundSetInfo = true
					break
				}
			}
		}
		if foundSetInfo {
			if strings.Contains(n.FirstChild.Data, "Encore") {
				if sl.Encore == nil {
					sl.Encore = &Set{}
				}
				getSongs(n, sl.Encore)
			} else {
				s := &Set{}
				sl.Sets = append(sl.Sets, s)
				getSongs(n, s)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSets(c)
		}

	}
	findSets(root)

	return sl, nil
}

func (s *Setlist) Songs() []string {
	var songs []string
	for _, set := range s.Sets {
		for _, song := range set.Songs {
			songs = append(songs, song)
		}
	}
	if s.Encore != nil {
		for _, song := range s.Encore.Songs {
			songs = append(songs, song)
		}
	}
	return songs
}

func (s *Setlist) String() string {
	str := fmt.Sprintf("ID{%s}DATE{%s}", s.ShowId, s.Date)
	for i, set := range s.Sets {
		str += fmt.Sprintf("SET%d{%s}", i+1, strings.Join(set.Songs, ","))
	}
	if s.Encore != nil {
		str += fmt.Sprintf("ENCORE{%s}", strings.Join(s.Encore.Songs, ","))
	}
	return str
}
