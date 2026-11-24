package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"
)

type RobotsChecker struct {
	client *http.Client
	mu     sync.Mutex
	cache  map[string]*robotstxt.RobotsData
}

func NewRobotsChecker(client *http.Client) *RobotsChecker {
	return &RobotsChecker{
		client: client,
		cache:  make(map[string]*robotstxt.RobotsData),
	}
}

func (r *RobotsChecker) Allowed(ctx context.Context, rawURL string) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false, fmt.Errorf("unsupported URL scheme")
	}

	if parsed.Hostname() == "" {
		return false, fmt.Errorf("missing hostname")
	}

	origin := parsed.Scheme + "://" + parsed.Host

	// Hold the lock through policy loading.
	// This prevents duplicate concurrent fetches.
	r.mu.Lock()
	defer r.mu.Unlock()

	rules, exists := r.cache[origin]

	if !exists {
		robotsURL := origin + "/robots.txt"

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			robotsURL,
			nil,
		)
		if err != nil {
			return false, err
		}

		resp, err := r.client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusNotFound ||
			resp.StatusCode == http.StatusGone:
			// No robots policy at this location.
			return true, nil

		case resp.StatusCode != http.StatusOK:
			// Conservative behavior: don't crawl if
			// the policy couldn't be established.
			return false, fmt.Errorf(
				"robots.txt returned HTTP %d",
				resp.StatusCode,
			)
		}

		// Bound the amount of robots.txt data we parse.
		const maxRobotsSize = 512 * 1024

		data, err := io.ReadAll(
			io.LimitReader(resp.Body, maxRobotsSize+1),
		)
		if err != nil {
			return false, err
		}

		if len(data) > maxRobotsSize {
			return false, fmt.Errorf("robots.txt too large")
		}

		rules, err = robotstxt.FromBytes(data)
		if err != nil {
			return false, fmt.Errorf(
				"robots.txt parsing failed: %w",
				err,
			)
		}

		r.cache[origin] = rules
	}

	group := rules.FindGroup("GoCrawler")

	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}

	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

	return group.Test(path), nil
}
