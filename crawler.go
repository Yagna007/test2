package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

func crawlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	customerType := r.FormValue("customerType")

	// Check if the page is cached and not expired
	isPageCached := false

	// Connect to the PostgreSQL database
	connStr := "user=group10 dbname=group_10 host=dpg-ckcm3t6smu8c73df2mug-a.oregon-postgres.render.com " +
		"password=ZGP4GRKuorYNj7yukOZlPBOV0JoiM6nw port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "Connection to database is successfull", http.StatusInternalServerError)
	}

	// defer db.Close()

	if isPageCached {
		// Serve the cached page
		fmt.Fprintln(w, "Cached Page Content")
	} else {
		// Real-time crawling logic here
		// Implement retry logic if the page is not available
		// Simulate storing the crawled data in the database
		if err := storeCrawledData(db, url, customerType); err != nil {
			http.Error(w, "Failed to store data in the database", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Real-Time Crawled Page Content")
	}
}

func storeCrawledData(db *sql.DB, url, customerType string) error {
	// Insert the crawled data into a database table
	query := "INSERT INTO crawled_data (url, customer_type, crawled_time) VALUES ($1, $2, $3)"
	_, err := db.Exec(query, url, customerType, time.Now())
	return err
}

// func main() {
// 	http.HandleFunc("/crawl", crawlHandler)
// 	http.ListenAndServe(":8080", nil)
// }
