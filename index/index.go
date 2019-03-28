package index

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/awbraunstein/setlist-search/index/query"
	"github.com/pkg/errors"
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

func (i *Index) ShowUrl(id string) string {
	sl := i.setlists[id]
	if sl != nil {
		return sl.Url
	}
	return ""
}

func (i *Index) Query(ctx context.Context, q string) ([]string, error) {
	p := query.NewParser(strings.NewReader(q))
	stmt, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return i.evaluate(ctx, stmt)
}

func (i *Index) evaluate(ctx context.Context, stmt query.Statement) ([]string, error) {
	var eval func(query.Statement) map[string]bool
	var err error
	eval = func(stmt query.Statement) map[string]bool {
		if deadline, ok := ctx.Deadline(); ok && deadline.After(time.Now()) {
			err = errors.New("Deadline exceeded for query")
			return nil
		}
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

	if err != nil {
		return nil, err
	}

	var showList []string
	for show := range shows {
		showList = append(showList, show)
	}

	sort.Strings(showList)
	return showList, nil

}
