package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/awbraunstein/gophish"
	"github.com/awbraunstein/setlist-search/index"
	"github.com/awbraunstein/setlist-search/searcher"
	"github.com/pkg/errors"
)

var usageMessage = `usage: indexer

indexer prepares the index used by the setlist-search app. The index is the file
named by $SETSEARCHERINDEX, or else $HOME/.setsearcherindex.


The indexer uses the phish.net api to scrape all of the new shows. If [-reset]
is false, then only new shows will be fetched.

The apikey for requests will be read from $PHISHAPIKEY.`

const (
	firstShowDate = "1983-10-30"
	bucketName    = "setlist-searcher-index"
	objectName    = "index.txt"
)

var (
	remote = flag.Bool("remote", true, "Whether the index will be stored remotely.")
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

func queryShowsGteDate(client *gophish.Client, lastShowDate string, showsFound map[int]*gophish.Show) error {
	resp, err := client.ShowsQuery(&gophish.ShowsQueryRequest{Order: "ASC", ShowdateGte: lastShowDate})
	if err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf("Request error: %d; %s", resp.ErrorCode, resp.ErrorMessage)
	}

	log.Printf("Querried %d shows.\n", resp.Response.Count)
	if resp.Response.Count == 0 {
		return nil
	}
	shows := resp.Response.Data
	lastShowFound, err := gophish.ParseDate(lastShowDate)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, show := range shows {
		showDateTime, err := gophish.ParseDate(show.ShowDate)
		if err != nil {
			return errors.WithStack(err)
		}
		if showDateTime.After(lastShowFound) {
			lastShowFound = showDateTime
		}
		if show.ArtistId != 1 {
			continue
		}
		log.Printf("Found show: %d, %s\n", show.ShowId, show.ShowDate)
		showsFound[show.ShowId] = show
	}

	if lastShowFoundStr := gophish.FormatDate(lastShowFound); lastShowFoundStr != lastShowDate {
		return queryShowsGteDate(client, lastShowFoundStr, showsFound)
	}
	return nil
}

func queryAllShows(client *gophish.Client) (map[int]*gophish.Show, error) {
	shows := make(map[int]*gophish.Show)
	if err := queryShowsGteDate(client, firstShowDate, shows); err != nil {
		return nil, err
	}
	return shows, nil
}

func getSetlistAndSongs(client *gophish.Client, show *gophish.Show) (*searcher.Setlist, map[string]string, error) {
	resp, err := client.SetlistsGet(&gophish.SetlistsGetRequest{ShowId: show.ShowId})
	if err != nil {
		return nil, nil, err
	}
	if resp.ErrorCode != 0 {
		return nil, nil, fmt.Errorf("Request error: %d; %s", resp.ErrorCode, resp.ErrorMessage)
	}

	if resp.Response.Count == 0 {
		// If we got 0 setlists, it means that there isn't a setlist for the given show.
		return nil, nil, nil
	}
	if resp.Response.Count > 1 {
		return nil, nil, fmt.Errorf("received multiple entries for showid=%d. Using the first one.", show.ShowId)
	}
	setlist := resp.Response.Data[0]
	return searcher.ParseSetlistFromPhishNet(setlist)
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

	apiKey := os.Getenv("PHISHAPIKEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Could not find api key $PHISHAPIKEY")
		usage()
	}
	client := gophish.NewClient(apiKey)

	indexLocation := getIndexLocation()
	w := index.NewWriter(indexLocation)

	shows, err := queryAllShows(client)
	if err != nil {
		log.Fatalf("error querying all shows: %v\n", err)
	}
	for _, show := range shows {
		sl, songSet, err := getSetlistAndSongs(client, show)
		if err != nil {
			log.Fatalf("unable to fetch setlist for show %d - %s; %v", show.ShowId, show.ShowDate, err)
		}
		if sl == nil {
			log.Printf("No known setlist for show %d - %s\n", show.ShowId, show.ShowDate)
			continue
		}
		w.AddSetlist(sl)
		for longName, shortName := range songSet {
			w.AddSong(longName, shortName)
		}
	}
	if err := w.Write(); err != nil {
		log.Fatalf("error writing file: %v\n", err)
	}
	log.Printf("wrote index to %s", indexLocation)

	// If this is remote, then we want to upload the result to Google Cloud Store.
	if *remote {
		f, err := os.Open(indexLocation)
		if err != nil {
			log.Fatalf("Unable to open index; %v\n", err)
		}
		defer f.Close()
		client, err := storage.NewClient(context.Background())
		if err != nil {
			log.Fatalf("Failed to create client: %v\n", err)
		}
		defer client.Close()
		object := client.Bucket(bucketName).Object(objectName)
		wc := object.NewWriter(context.Background())
		if _, err = io.Copy(wc, f); err != nil {
			log.Fatalf("Unable to copy index; %v\n", err)
		}
		if err := wc.Close(); err != nil {
			log.Fatalf("Unable to close index; %v\n", err)
		}
		log.Printf("wrote index to the cloud at %s/%s", bucketName, objectName)
	}
}
