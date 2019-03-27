package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	queryShowsUrl = "https://api.phish.net/v3/shows/query"
	getSetlistUrl = "https://api.phish.net/v3/setlists/get"

	dateFormat = "2006-01-02"
	queryRate  = time.Minute / 120
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

var (
	apiKey = os.Getenv("PHISHAPIKEY")
)

type showData struct {
	date string
	id   string
}

type httpResult struct {
	resp *http.Response
	err  error
}

var (
	throttle = time.Tick(queryRate)
	//httpResults  = make(chan httpResult)
	//httpRequests = make(chan *http.Request)
)

func sendPhishNetQuery(url string) (*http.Response, error) {
	log.Printf("Querying: %s\n", url)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	<-throttle
	return http.DefaultClient.Do(req)
}

func queryShowsGteDate(lastShowDate string, showsFound map[string]showData) error {
	url := queryShowsUrl + "?apikey=" + apiKey
	url += "&order=ASC"
	url += "&showdate_gte=" + lastShowDate
	res, err := sendPhishNetQuery(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var respJson map[string]interface{}
	if err := json.Unmarshal(body, &respJson); err != nil {
		return errors.WithStack(err)
	}

	errorCode, ok := respJson["error_code"].(float64)
	if !ok {
		return errors.New("Unable to find error code")
	}
	if errorCode != 0 {
		return fmt.Errorf("Request error: %d; %s", int(errorCode), respJson["error_message"].(string))
	}

	response := respJson["response"].(map[string]interface{})
	count := int(response["count"].(float64))
	log.Printf("Querried %d shows.\n", count)
	if count == 0 {
		return nil
	}
	shows := response["data"].([]interface{})
	lastShowFound, err := time.Parse(dateFormat, lastShowDate)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, show := range shows {
		show := show.(map[string]interface{})
		showDate := show["showdate"].(string)
		showDateTime, err := time.Parse(dateFormat, showDate)
		if err != nil {
			return errors.WithStack(err)
		}
		if showDateTime.After(lastShowFound) {
			lastShowFound = showDateTime
		}
		if int(show["artistid"].(float64)) != 1 {
			continue
		}
		showId := strconv.FormatInt(int64(show["showid"].(float64)), 10)
		log.Printf("Found show: %s, %s\n", showId, showDate)
		showsFound[showId] = showData{
			date: showDate,
			id:   showId,
		}
	}

	if lastShowFoundStr := lastShowFound.Format(dateFormat); lastShowFoundStr != lastShowDate {
		return queryShowsGteDate(lastShowFoundStr, showsFound)
	}
	return nil
}

func queryAllShows() (map[string]showData, error) {
	shows := make(map[string]showData)
	if err := queryShowsGteDate(firstShowDate, shows); err != nil {
		return nil, err
	}
	return shows, nil
}

func getSetlistAndSongs(showId, date string) (*searcher.Setlist, map[string]string, error) {
	url := fmt.Sprintf("%s?apikey=%s&showid=%s", getSetlistUrl, apiKey, showId)
	res, err := sendPhishNetQuery(url)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	var respJson map[string]interface{}
	if err := json.Unmarshal(body, &respJson); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	errorCode, ok := respJson["error_code"].(float64)
	if !ok {
		return nil, nil, errors.New("Unable to find error code")
	}
	if errorCode != 0 {
		return nil, nil, fmt.Errorf("Request error: %d; %s", int(errorCode), respJson["error_message"].(string))
	}

	response := respJson["response"].(map[string]interface{})
	count := int(response["count"].(float64))
	if count == 0 {
		// If we got 0 setlists, it means that there isn't a setlist for the given show.
		return nil, nil, nil
	}
	if count > 1 {
		return nil, nil, fmt.Errorf("received multiple entries for showid=%s. Using the first one.", showId)
	}
	data := response["data"].([]interface{})[0].(map[string]interface{})
	setlistData := data["setlistdata"].(string)
	return searcher.ParseSetlistFromPhishNet(showId, date, data["url"].(string), setlistData)
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

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Could not find api key $PHISHAPIKEY")
		usage()
	}

	indexLocation := getIndexLocation()
	w := index.NewWriter(indexLocation)

	shows, err := queryAllShows()
	if err != nil {
		log.Fatalf("error querying all shows: %v\n", err)
	}
	for _, show := range shows {
		sl, songSet, err := getSetlistAndSongs(show.id, show.date)
		if err != nil {
			log.Fatalf("unable to fetch setlist for show %s - %s; %v", show.id, show.date, err)
		}
		if sl == nil {
			log.Printf("No known setlist for show %s - %s\n", show.id, show.date)
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
}
