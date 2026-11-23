package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"errors"
)

var ErrNonHTML = errors.New("non-HTML content")

type HTTPFetcher struct {
	Client *http.Client
}

func (h HTTPFetcher) Fetch(url string) (body string, urls []string, err error) {
	resp, err := h.Client.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	//checking http status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	// only wan tto acccept html pages
	contentType := resp.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return "", nil, err
	}
	if mediaType != "text/html" {
    return "", nil, ErrNonHTML
	}

	//don't want to exceed 2 MiB
	const maxBodySize = 2 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return "", nil, err
	}
	if len(data) > maxBodySize {
		return "", nil, fmt.Errorf("response body too large: %d bytes", len(data))
	}
	body = string(data)

	urls, err = extractLinks(body, resp.Request.URL.String())
	if err != nil {
		return "", nil, err
	}

	return body, urls, nil
}

func extractLinks(body string, pageURL string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	var links []string
	seen := make(map[string]bool)

	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				href := strings.TrimSpace(attr.Val)

				if href == "" || strings.HasPrefix(href, "#") {
					continue
				}

				link, err := url.Parse(href)
				if err != nil {
					continue
				}

				absoluteURL := baseURL.ResolveReference(link)

				if absoluteURL.Scheme != "http" &&
					absoluteURL.Scheme != "https" {
					continue
				}

				if absoluteURL.Hostname() == "" {
					continue
				}

				absoluteURL.Fragment = ""
				absoluteURL.RawFragment = ""

				normalized := absoluteURL.String()

				if !seen[normalized] {
					seen[normalized] = true
					links = append(links, normalized)
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	return links, nil
}
