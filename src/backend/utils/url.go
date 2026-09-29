package utils

import (
	"errors"
	"net/url"
)

// SanitizeURL validates an http(s) URL and strips playlist parameters so
// only the linked video is downloaded.
func SanitizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("url must be an http(s) link")
	}

	q := u.Query()
	if q.Has("list") || q.Has("index") {
		q.Del("list")
		q.Del("index")
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}
