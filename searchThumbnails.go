package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// For unmarshalling the JSON response
type imageData struct { 
	ResponseData struct {
		Results []Thumbnail `json:"results"`
	} `json:"responseData"`
}

type Thumbnail struct {
	Url string `json:"tbUrl"`
}

func main() {

	// arg processing
	search := flag.String("s", "", "search keywords")
	numResults := flag.Int("n", 4, "number of results")
	flag.Parse()

	// Create a URL channel
	cUrl := make(chan string)
	go GetThumbnails(*search, *numResults, cUrl)

	// Download the images
	DownloadImages(*search, *numResults, cUrl)
	
	// Finish
	log.Println("Done")

}

// GetSearchURL
// Constructs a Google Search API URL for the given keywords
func GetURLBase(search string) string {

	// formatter for a Google Image search
	urlBase := "http://ajax.googleapis.com/ajax/services/search/images?v=1.0&q=%s&start="

    // construct the url
	query := strings.Replace(search, " ", "+", -1)
	return fmt.Sprintf(urlBase, query)	
}

// GetThumbnails
// Query the Google Image API for numResults for the search
func GetThumbnails(search string, numResults int, cUrl chan string) {

	var url string

	// Get the urlBase
	urlBase := GetURLBase(search)

	// Google API only returns 4 results per query
	var numIter int = int(math.Ceil(float64(numResults)/ 4))

	// Create work groups
	var wg sync.WaitGroup
	wg.Add(numIter)

	log.Printf("Searching Google Images for '%s'\n", search)
	
	for i := 0; i < numIter; i++ {

		url = urlBase + strconv.Itoa(i)

		// Goroutine for each request
		go func (url string) {
			// Defer close for the WaitGroup
			defer wg.Done()

			// tmp var for unmarshalled JSON
			var tmp imageData

			// GET images
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal("http.Get: ", err)
			}
			defer resp.Body.Close()

			// Decode to tmp
			if err = json.NewDecoder(resp.Body).Decode(&tmp); err != nil {
				log.Fatal("Decoder.Decode: ", err)
			}

			// pass URLs onto the channel
			for _, result := range tmp.ResponseData.Results {
				cUrl <- result.Url
			} 
		} (url)
	}

	wg.Wait() // wait until the url checks complete
	close(cUrl)
}

// DownloadImages
// Download all of the image URLs in thumbs to a new directory
func DownloadImages(search string, numResults int, cUrl chan string) {

	var i int = 0

	// Time layout formatter
	const layout = "_Jan2_06_3:04"

	// Construct a path and filename
	filenameBase := strings.Replace(search, " ", "_", -1)
	dir := filenameBase + time.Now().Local().Format(layout)

	// Make the directory for the images
	if err := os.Mkdir(dir, 0755); err != nil {
		log.Fatal("os.Mkdir: ", err)
	}

	// Download each URL
	log.Printf("Downloading thumbnails to '%s'\n", dir)

	// Create work groups
	var wg sync.WaitGroup
	wg.Add(numResults)

	for url := range cUrl {

		go func (url string, i int) {
			// Defer close for the WaitGroup
			defer wg.Done()

			// GET images
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal("http.Get: ", err)
			}
			defer resp.Body.Close()

	    	// Read the image response
		    data, err := ioutil.ReadAll(resp.Body)
		    if err != nil {
		        log.Fatal("ioutil.ReadAll: ", err)
		    }

		    // Write file to the directory
	    	ioutil.WriteFile(filepath.Join(dir, filenameBase + strconv.Itoa(i) + `.jpg`), data, 0755)
		} (url, i)

		// increment i
		i++
	}

	wg.Wait() // wait until the url checks complete
}