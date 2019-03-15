package index

import (
	"sort"
	"strings"

	"github.com/awbraunstein/setlist-search/index/query"
)

func (i *Index) Songs() map[string]string {
	return i.songs
}

func (i *Index) ShowDate(id string) string {
	sl := i.setlists[id]
	if sl != nil {
		return sl.Date
	}
	return ""
}

func (i *Index) Query(q string) ([]string, error) {
	p := query.NewParser(strings.NewReader(q))
	stmt, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return i.evaluate(stmt), nil
}

func (i *Index) evaluate(stmt query.Statement) []string {
	var eval func(query.Statement) map[string]bool
	eval = func(stmt query.Statement) map[string]bool {
		switch n := stmt.(type) {
		case *query.AndStatement:
			leftShows := eval(n.Left)
			rightShows := eval(n.Right)
			// intersection
			newShows := make(map[string]bool)
			for show := range leftShows {
				if rightShows[show] {
					newShows[show] = true
				}
			}
			return newShows
		case *query.OrStatement:
			leftShows := eval(n.Left)
			rightShows := eval(n.Right)
			// union
			for show := range leftShows {
				rightShows[show] = true
			}
			return rightShows
		case *query.NotStatement:
			shows := eval(n.S)
			newShows := make(map[string]bool)
			for show := range i.setlists {
				if !shows[show] {
					newShows[show] = true
				}
			}
			return newShows
		case *query.Expression:
			shows := make(map[string]bool)
			for _, show := range i.reverseIndex[n.Value] {
				shows[show] = true
			}
			return shows
		}
		return nil

	}

	shows := eval(stmt)

	var showList []string
	for show := range shows {
		showList = append(showList, show)
	}

	sort.Strings(showList)
	return showList

}
