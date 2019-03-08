package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/awbraunstein/setlist-search/index"
)

var usageMessage = `usage: searcher

searcher opens the index used by the setlist-search app. The index is the file
named by $SETSEARCHERINDEX, or else $HOME/.setsearcherindex. It then allows for
multiple queries on the index.
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

func getIndexLocation() string {
	if indexLocation := os.Getenv("SETSEARCHERINDEX"); indexLocation != "" {
		return indexLocation
	}
	return filepath.Clean(os.Getenv("HOME") + "/.setsearcherindex")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	idxLoc := getIndexLocation()
	fmt.Printf("Opening index: %s\n", idxLoc)
	start := time.Now()
	i, err := index.Open(idxLoc)
	if err != nil {
		log.Fatalf("Unable to open index; %v\n", err)
	}
	fmt.Printf("Took %v to open index\n", time.Since(start))
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := scanner.Text()
		start := time.Now()
		shows, err := i.Query(line)
		fmt.Printf("Query took %v\n", time.Since(start))
		if err != nil {
			fmt.Printf("Error processing query: %v\n", err)
			continue
		}
		var dates []string
		for _, show := range shows {
			dates = append(dates, i.ShowDate(show))
		}
		sort.Strings(dates)

		fmt.Printf("Matched shows %d:\n%v\n", len(dates), strings.Join(dates, ", "))
		fmt.Print("> ")
	}
	fmt.Println("Goodbye")
}
