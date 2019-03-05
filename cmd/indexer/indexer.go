package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

var usageMessage = `usage: indexer [-reset]

indexer prepares the index used by the setlist-search app. The index is the file
named by $SETSEARCHERINDEX, or else $HOME/.setsearcherindex.


The indexer uses the phish.net api to scrape all of the new shows. If [-reset]
is false, then only new shows will be fetched.

The apikey for requests will be read from $PHISHAPIKEY.

The -reset flag will re-fetch all shows and rebuild the index from scratch.
`

const (
	firstShowDate = "1983-10-30"
	queryShowsUrl = "https://api.phish.net/v3/shows/query"
	dateFormat    = "2006-01-02"
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

var (
	resetFlag   = flag.Bool("reset", false, "discard existing index")
	verboseFlag = flag.Bool("verbose", false, "print extra information")
)

type showData struct {
	showDate string
	showId   int
}

func queryShowsGteDate(apiKey string, lastShowDate string, showsFound map[int]showData) error {
	url := queryShowsUrl + "?apikey=" + apiKey
	url += "&order=ASC"
	url += "&showdate_gte=" + lastShowDate
	fmt.Printf("Querying: %s\n", url)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
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
		return fmt.Errorf("Request error: %d; %s", errorCode, respJson["error_message"].(string))
	}

	response := respJson["response"].(map[string]interface{})
	count := int(response["count"].(float64))
	fmt.Printf("Querried %d shows.\n", count)
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
		showId := int(show["showid"].(float64))
		fmt.Printf("Found show: %d, %s\n", showId, showDate)
		showsFound[showId] = showData{
			showDate: showDate,
			showId:   showId,
		}
	}

	if lastShowFoundStr := lastShowFound.Format(dateFormat); lastShowFoundStr != lastShowDate {
		return queryShowsGteDate(apiKey, lastShowFoundStr, showsFound)
	}
	return nil
}

func queryAllShows(apiKey string) (map[int]showData, error) {
	shows := make(map[int]showData)
	if err := queryShowsGteDate(apiKey, firstShowDate, shows); err != nil {
		return nil, err
	}
	return shows, nil
}

func main() {
	flag.Usage = usage
	flag.Parse()

	apiKey := os.Getenv("PHISHAPIKEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Could not find api key $PHISHAPIKEY")
		usage()
	}

	if *resetFlag {
		shows, err := queryAllShows(apiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying all shows: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("Found %d shows.\n", len(shows))
	}
}
