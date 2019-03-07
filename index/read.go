package index

// Index is the setlist searcher index that's loaded into memory to run analysis
// on setlists.
type Index struct {
}

// Open reads an Index from disk.
func Open(filename string) *Index {
	return &Index{}
}
