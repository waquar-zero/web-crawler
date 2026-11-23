package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {
	//take input from command line
	urlFlag := flag.String("url", "", "Starting URL to Crawl")
	depthFlag := flag.Int("depth", 3, "Maximum Crawl Depth")
	sameDomainFlag := flag.Bool(
		"same-domain",
		true,
		"Only crawl URLs belonging to the starting hostname",
	)

	//parse the flags
	flag.Parse()

	//Validation
	if *urlFlag == "" {
		fmt.Println("Please provide a starting URL using the -url flag.")
		return
	}

	parsedURL, err := url.Parse(*urlFlag)
	if err != nil {
		fmt.Println("Invalid URL:", err)
		return
	}

	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
		parsedURL.Hostname() == "" {
		fmt.Println("Error: URL must begin with http:// or https:// and contain a hostname")
		return
	}

	if *depthFlag <= 0 {
		fmt.Println("Error: depth must be greater than zero")
		return
	}
	if *depthFlag > 5 {
		fmt.Println("Error: maximum supported depth is 5")
		return
	}

	//creating the HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	fetcher := HTTPFetcher{Client: client}

	//crawl beings
	fmt.Printf("Starting crawl: %s (depth: %d)\n", *urlFlag, *depthFlag)

	Crawl(*urlFlag, *depthFlag, fetcher, *sameDomainFlag)

	fmt.Println("Crawl completed.")
}
