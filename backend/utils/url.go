package utils

import (
	"net/url"
)

// SanitizeURL removes the 'list' parameter from a URL to avoid downloading entire playlists
func SanitizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	if q.Has("list") || q.Has("index") {
		q.Del("list")
		q.Del("index")
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}
