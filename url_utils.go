package main

import (
	"net"
	"net/url"
	"strings"
	"path"
)

func shouldSkipURL(rawURL string) bool {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        return true
    }

    ext := strings.ToLower(path.Ext(parsed.Path))

    blockedExtensions := map[string]bool{
        ".pdf":  true,
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".gif":  true,
        ".webp": true,
        ".svg":  true,
        ".mp4":  true,
        ".mp3":  true,
        ".zip":  true,
        ".exe":  true,
        ".css":  true,
        ".js":   true,
    }

    return blockedExtensions[ext]
}

func normalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Hostname())

	port := parsed.Port()
	if (parsed.Scheme == "https" && port == "443") || (parsed.Scheme == "http" && port == "80") {

		port = ""
	}

	if port != "" {
		parsed.Host = net.JoinHostPort(hostname, port)
	} else {
		parsed.Host = hostname
	}
	parsed.Fragment = ""
	parsed.RawFragment = ""
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String(), nil
}
