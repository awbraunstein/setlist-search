package searcher

import (
	"strconv"
)

// Searcher is the result of a compiled query. A Searcher is safe for concurrent
// use by multiple goroutines.
type Searcher struct {
	expr string // as passed to Compile
}

// Compile parses a searcher query and returns, if successful, a Searcher that
// can be used to match against setlists.
func Compile(expr string) (*Searcher, error) {
	searcher := &Searcher{
		expr: expr,
	}
	return searcher, nil
}

// MustCompile is like Compile but panics if expression cannot be parsed. It
// simplifies safe initialization of global variables holding compiled
// searchers.
func MustCompile(expr string) *Searcher {
	s, err := Compile(expr)
	if err != nil {
		panic(`searcher: Compile(` + quote(expr) + `): ` + err.Error())
	}
	return s
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

// FindShows looks through the list of shows and returns the matching show ids.
func (s *Searcher) FindShows(shows string) []string {
	return nil
}
