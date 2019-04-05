package searcher

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/awbraunstein/gophish"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Setlist holds the set from a show.
type Setlist struct {
	ShowId int
	Date   string
	Sets   []*Set
	Encore *Set
	Url    string
}

// Set holds a single set of a setlist.
type Set struct {
	Songs []string
}

var (
	idRe     = regexp.MustCompile(`^ID\{(\d+?)\}`)
	dateRe   = regexp.MustCompile(`DATE\{(.+?)\}`)
	urlRe    = regexp.MustCompile(`URL\{(.+?)\}`)
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
	id, err := strconv.Atoi(idMatches[1])
	if err != nil {
		return nil, fmt.Errorf("ParseSetlist: couldn't parse id into an int %s", idMatches[1])
	}
	dateMatches := dateRe.FindStringSubmatch(setlist)
	if len(dateMatches) != 2 {
		return nil, fmt.Errorf("ParseSetlist: couldn't find DATE tag in setlist: %s", setlist)
	}
	urlMatches := urlRe.FindStringSubmatch(setlist)
	if len(urlMatches) != 2 {
		return nil, fmt.Errorf("ParseSetlist: couldn't find URL tag in setlist: %s", setlist)
	}
	sl := &Setlist{
		ShowId: id,
		Date:   dateMatches[1],
		Url:    urlMatches[1],
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

// normalizeName returns the normalized song name.
// A Name Like This -> a-name-like-this
func normalizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsUpper(r) {
			return unicode.ToLower(r)
		}
		switch r {
		case '.', ',', ';', ':', '\'':
			return -1
		case ' ':
			return '-'
		}
		return r
	}, name)
}

// Returns a setlist and the songset or an error if there were any.
func ParseSetlistFromPhishNet(setlist *gophish.Setlist) (*Setlist, map[string]string, error) {
	sl := &Setlist{
		ShowId: setlist.ShowId,
		Date:   setlist.ShowDate,
		Url:    setlist.Url,
	}
	root, err := html.Parse(strings.NewReader(setlist.SetlistData))
	if err != nil {
		return nil, nil, err
	}

	songSet := make(map[string]string)

	var getSongsErr error

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
				var humanName string
				for _, attr := range current.Attr {
					if attr.Key == "href" {
						// if this is a song href, then
						// there should be a single
						// child text node.
						child := current.FirstChild
						if child == nil || child.Type != html.TextNode {
							getSongsErr = fmt.Errorf("expected node %v to have a child text node", current)
							return
						}
						humanName = strings.TrimSpace(child.Data)
						name = normalizeName(humanName)
						break
					}
				}
				if name == "" || humanName == "" {
					getSongsErr = fmt.Errorf("Expected node %v to be a song node", current)
					return
				}
				set.Songs = append(set.Songs, name)
				songSet[humanName] = name
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

	if getSongsErr != nil {
		return nil, nil, getSongsErr
	}

	return sl, songSet, nil
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
	str := fmt.Sprintf("ID{%d}DATE{%s}URL{%s}", s.ShowId, s.Date, s.Url)
	for i, set := range s.Sets {
		str += fmt.Sprintf("SET%d{%s}", i+1, strings.Join(set.Songs, ","))
	}
	if s.Encore != nil {
		str += fmt.Sprintf("ENCORE{%s}", strings.Join(s.Encore.Songs, ","))
	}
	return str
}
