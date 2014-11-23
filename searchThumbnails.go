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

	// Get thumbnail URLs
	thumbs := GetThumbnails(*search, *numResults)

	// Download the images
	DownloadImages(*search, &thumbs)
	
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
func GetThumbnails(search string, numResults int) []string {

	var resp *http.Response
	var err error
	var url string

	// Get the urlBase
	urlBase := GetURLBase(search)

	// Google API only returns 4 results per query
	var numIter int = int(math.Ceil(float64(numResults)/ 4))

	// Create thumbnail arrays
	thumbs := make([]string, numIter * 4)
	var tmp imageData

	log.Printf("Searching Google Images for '%s'\n", search)
	for i := 0; i < numIter; i++ {

		url = urlBase + strconv.Itoa(i)

		// GET images
		resp, err = http.Get(url)
		if err != nil {
			log.Fatal("http.Get: ", err)
		}
		defer resp.Body.Close()

		// Decode to tmp
		if err = json.NewDecoder(resp.Body).Decode(&tmp); err != nil {
			log.Fatal("Decoder.Decode: ", err)
		}

		// fill thumbs with tmp
		for j := 0; j < 4; j++ {
			thumbs[j + (i * 4)] = tmp.ResponseData.Results[j].Url
		}
	}

	return thumbs[0:numResults]
}

// DownloadImages
// Download all of the image URLs in thumbs to a new directory
func DownloadImages(search string, thumbs *[]string) {

	var resp *http.Response
	var err error
	var data []byte

	// Time layout formatter
	const layout = "_Jan2_06_3:04"

	// Construct a path and filename
	filenameBase := strings.Replace(search, " ", "_", -1)
	dir := filenameBase + time.Now().Local().Format(layout)

	// Make the directory for the images
	if err = os.Mkdir(dir, 0755); err != nil {
		log.Fatal("os.Mkdir: ", err)
	}

	// Download each URL
	log.Printf("Downloading thumbnails to '%s'\n", dir)

	for i, url := range *thumbs {

		// GET images
		resp, err = http.Get(url)
		if err != nil {
			log.Fatal("http.Get: ", err)
		}
		defer resp.Body.Close()

    	// We read all the bytes of the image
    	// Types: data []byte
	    data, err = ioutil.ReadAll(resp.Body)
	    if err != nil {
	        log.Fatal("ioutil.ReadAll: ", err)
	    }

	    // Write file to the directory
	    ioutil.WriteFile(filepath.Join(dir, filenameBase + strconv.Itoa(i) + `.jpg`), data, 0755)
	}
}