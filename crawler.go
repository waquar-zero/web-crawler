package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) (
		body string,
		urls []string,
		err error,
	)
}

func isSameDomain(link string, startingHost string) bool {
	parsed, err := url.Parse(link)
	if err != nil {
		return false
	}
	return parsed.Hostname() == startingHost
}

func crawlHelper(
	ctx context.Context,
	url string,
	depth int,
	f Fetcher,
	m map[string]bool,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	sem chan struct{},
	startingHost string,
	sameDomainOnly bool,
	maxPages int,
	robots *RobotsChecker,
	limiter *HostLimiter,
) {
	defer wg.Done()
	
	if depth <= 0 || ctx.Err() != nil {
    return
	}

	//normalization of url
	normalizedURL, err := normalizeURL(url)
	if err != nil {
		fmt.Printf("Invalid URL %s: %v\n", url, err)
		return
	}

	url = normalizedURL

	// Step 1: Check whether URL was visited
	mu.Lock()

	if m[url] {
		mu.Unlock()
		return
	}

	if len(m) >= maxPages {
		mu.Unlock()
		return
	}

	m[url] = true
	mu.Unlock()

	
	allowed, err := robots.Allowed(ctx, url)
	if err != nil {
		fmt.Printf("Skipping %s: robots check failed: %v\n", url, err)
		return
	}

	if !allowed {
		fmt.Printf("Skipping %s: disallowed by robots.txt\n", url)
		return
	}

	if err := limiter.Wait(ctx, url); err != nil {
		return
	}

	// Step 2: Acquire a semaphore slot
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	_, urls, err := f.Fetch(ctx, url)

	// Release the slot after fetching
	<-sem

	if err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return
		}

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

	if ctx.Err() != nil {
		return
	}
	// Step 5: Crawl discovered URLs
	for _, u := range urls {

		if ctx.Err() != nil {
			return
		}

		if shouldSkipURL(u) {
			continue
		}

		if sameDomainOnly && !isSameDomain(u, startingHost) {
			continue
		}
		wg.Add(1)

		go crawlHelper(
			ctx,
			u,
			depth-1,
			f,
			m,
			mu,
			wg,
			sem,
			startingHost,
			sameDomainOnly,
			maxPages,
			robots,
			limiter,
		)
	}
}

func Crawl(
	ctx context.Context,
	startURL string,
	depth int,
	f Fetcher,
	sameDomainOnly bool,
	maxPages int,
	robots *RobotsChecker,
	limiter *HostLimiter,
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
		ctx,
		startURL,
		depth,
		f,
		visited,
		&mu,
		&wg,
		sem,
		startingHost,
		sameDomainOnly,
		maxPages,
		robots,
		limiter,
	)

	wg.Wait()
}
