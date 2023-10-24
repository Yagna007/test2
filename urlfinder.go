package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func finder(w http.ResponseWriter, r *http.Request) {

	// url we want to crawl
	url := r.FormValue("url")
	customerType := r.FormValue("customerType")
	fmt.Println(customerType)

	if url == "" {
		http.Error(w, "Missing or invalid 'url' parameter", http.StatusBadRequest)
		return
	}

	startURL := url

	// Create a slice to store unique URLs
	visitedUrlTime := make(map[string]bool)
	var queue []string

	// Add the starting URL to the queue
	queue = append(queue, startURL)

	// Define the maximum number of URLs to crawl
	maxURLsToCrawl := 10

	// Create an HTTP client
	client := &http.Client{}
	cnt := 0

	for len(queue) > 0 && len(visitedUrlTime) < maxURLsToCrawl {
		// Dequeue the first URL
		url := queue[0]
		queue = queue[1:]

		// Check if the URL has already been visited
		if visitedUrlTime[url] {
			continue
		}

		// Fetch the HTML content of the page
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error fetching %s: %s", url, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error: %s returned status code %d", url, resp.StatusCode)
			continue
		}

		// Parse the HTML document
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("Error parsing HTML from %s: %s", url, err)
			continue
		}

		// Add the URL to the list of visited URLs
		visitedUrlTime[url] = true

		// Extract and print all the URLs on the page
		doc.Find("a").Each(func(index int, item *goquery.Selection) {
			href, _ := item.Attr("href")
			absURL := resolveURL(url, href)
			fmt.Println(fmt.Sprintf("%d %s", cnt, absURL))
			cnt = cnt + 1
			// Add the absolute URL to the queue for further crawling
			queue = append(queue, absURL)
		})
	}
}

func resolveURL(base, relative string) string {
	u, err := url.Parse(relative)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	absURL := baseURL.ResolveReference(u)
	return absURL.String()
}

// func main() {
// 	http.HandleFunc("/", indexHandler)
// 	http.HandleFunc("/crawl", finder)
// 	http.ListenAndServe(":8080", nil)
// }
