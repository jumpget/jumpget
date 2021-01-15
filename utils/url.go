package utils

import (
	"net/url"
	"strings"
)

func IsValidURL(u string) bool {
	u = strings.ToLower(u)
	if !(strings.HasPrefix(u, "http") || strings.HasPrefix(u, "https")) {
		return false
	}

	_, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}

	parsed, err := url.Parse(u)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return true
}
