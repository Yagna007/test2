package main

import (
	"container/heap"
	"fmt"
	"log"
	// "os"
	// "io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	// "github.com/gocolly/colly/v2"
	"github.com/PuerkitoBio/goquery"
)

type URLItem struct {
	URL      string
	Priority int
	wclient http.ResponseWriter
}

type PriorityQueue []*URLItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority // Higher priority first
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*URLItem)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

var mu sync.Mutex
var lastCrawlTime time.Time

func indexHandler2(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func finder2(w http.ResponseWriter, r *http.Request) {
	// URL and customer type handling...

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
	visitedUrlTime := make(map[string]time.Time)
	subUrlInCurrentUrl := make(map[string][]string)

	// Initialize a priority queue
	priorityQueue := make(PriorityQueue, 0)
	heap.Init(&priorityQueue)

	// Add the starting URL with appropriate priority
	priority := 1
	if customerType == "paying" {
		priority = 2 // Higher priority for paid users
	}
	item := &URLItem{
		URL:      startURL,
		Priority: priority,
		wclient:        w,
	}
	heap.Push(&priorityQueue, item)

	// Define the maximum number of URLs to crawl
	maxURLsToCrawl := 100

	// Create an HTTP client
	client := &http.Client{}
	cnt := 0

	for priorityQueue.Len() > 0 && len(visitedUrlTime) < maxURLsToCrawl {
		mu.Lock()
		item := heap.Pop(&priorityQueue).(*URLItem)
		mu.Unlock()

		url := item.URL
		priority := item.Priority
		writerClient := item.wclient

		// Check if the URL has already been visited
		crawlTime, exists := visitedUrlTime[url]

		if exists {
			if time.Since(crawlTime) < 60*time.Minute {
				// dont visit again
				if values, exists := subUrlInCurrentUrl[url]; exists {
					for _, value := range values {
						fmt.Fprintln(writerClient,value)
					}
				} else {
					fmt.Printf("Key %s not found in the map\n", url)
				}
				// return the stored links with respect to this
				continue
			} else {
				delete(visitedUrlTime, url)
				delete(subUrlInCurrentUrl, url)
			}
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
		visitedUrlTime[url] = time.Now()

		// Add the absolute URL to the queue with the next priority
		doc.Find("a").Each(func(index int, item *goquery.Selection) {
			href, _ := item.Attr("href")
			absURL := resolveURL2(url, href)

			// mutex lock
			mu.Lock()

			fmt.Fprintln(writerClient,fmt.Sprintf("%d %s %d", cnt, absURL, priority))
			cnt = cnt + 1

			// heap.Push(&priorityQueue, &URLItem{
			// 	URL:      absURL,
			// 	Priority: priority,
			// 	wclient:  writerClient,
			// })

			// map of url , sub urls in the same page.
			subUrlInCurrentUrl[url] = append(subUrlInCurrentUrl[url], absURL)

			mu.Unlock()
		})
	}
}

func resolveURL2(base, relative string) string {
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

func main() {
	http.HandleFunc("/", indexHandler2)
	http.HandleFunc("/crawl", finder2)
	http.ListenAndServe(":8080", nil)
}


