package main

import (
	"fmt"
	"net/url"
	"sync"
	"errors"
)

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

func isSameDomain(link string, startingHost string) bool {
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parsed.Hostname() == startingHost
}

func crawlHelper(
	url string,
	depth int,
	f Fetcher,
	m map[string]bool,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	sem chan struct{},
	startingHost string,
	sameDomainOnly bool,
) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	// Step 1: Check whether URL was visited
	mu.Lock()

	if m[url] {
		mu.Unlock()
		return
	}

	m[url] = true
	mu.Unlock()

	// Step 2: Acquire a semaphore slot
	sem <- struct{}{}

	// Step 3: Fetch the URL
	_, urls, err := f.Fetch(url)

	// Step 4: Release the semaphore slot
	<-sem

	if err != nil {
    if errors.Is(err, ErrNonHTML) {
        return
    }

    fmt.Printf("Failed to crawl %s: %v\n", url, err)
    return
	}

	fmt.Printf(
		"Crawled: %s | Links found: %d\n",
		url,
		len(urls),
	)

	if depth == 1 {
		return
	}

	// Step 5: Crawl discovered URLs
	for _, u := range urls {

		// wont be checking external websites if sameDomainOnly is true
		if sameDomainOnly && !isSameDomain(u, startingHost) {
			continue
		}

		wg.Add(1)

		go crawlHelper(
			u,
			depth-1,
			f,
			m,
			mu,
			wg,
			sem,
			startingHost,
			sameDomainOnly,
		)
	}
}

func Crawl(
	startURL string,
	depth int,
	f Fetcher,
	sameDomainOnly bool,
) {
	visited := make(map[string]bool)

	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 5)

	// Extract the starting hostname
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		fmt.Println("Invalid starting URL:", err)
		return
	}

	startingHost := parsedURL.Hostname()

	wg.Add(1)

	crawlHelper(
		startURL,
		depth,
		f,
		visited,
		&mu,
		&wg,
		sem,
		startingHost,
		sameDomainOnly,
	)

	wg.Wait()
}
