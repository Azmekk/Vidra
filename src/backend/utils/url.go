package utils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	urlInText     = regexp.MustCompile(`(?i)\bhttps?://[^\s<>"'` + "`" + `]+`)
	bareURLInText = regexp.MustCompile(`(?i)\b(?:www\.)?[a-z0-9-]+(?:\.[a-z0-9-]+)*\.[a-z]{2,}/[^\s<>"'` + "`" + `]*`)
)

// ExtractURL pulls the first link out of shared text such as
// "Check this out on AppName: https://…  Follow us!".
func ExtractURL(text string) string {
	text = strings.TrimSpace(text)
	if m := urlInText.FindString(text); m != "" {
		return trimTrailing(m)
	}
	if m := bareURLInText.FindString(text); m != "" {
		return "https://" + trimTrailing(m)
	}
	return text
}

func trimTrailing(s string) string {
	s = strings.TrimRight(s, ".,;:!?…»”’")
	for _, pair := range [][2]string{{"(", ")"}, {"[", "]"}, {"{", "}"}} {
		for strings.HasSuffix(s, pair[1]) && strings.Count(s, pair[0]) < strings.Count(s, pair[1]) {
			s = strings.TrimSuffix(s, pair[1])
		}
	}
	return strings.TrimRight(s, ".,;:!?…»”’")
}

// SanitizeURL extracts an http(s) link from text and strips playlist
// parameters so only the linked video is downloaded.
func SanitizeURL(raw string) (string, error) {
	u, err := url.Parse(ExtractURL(raw))
	if err != nil {
		return "", err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("no http(s) link found")
	}

	q := u.Query()
	if q.Has("list") || q.Has("index") {
		q.Del("list")
		q.Del("index")
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}
