package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/div197/bob-gemini-free/internal/config"
)

// keyMayTravelTo reports whether a gateway API key may be sent to targetURL.
// A key goes to HTTPS targets, or over HTTP only to a loopback address. Plain
// HTTP to any other host would expose the key on the network.
func keyMayTravelTo(targetURL string) bool {
	u, err := url.Parse(targetURL)
	if err != nil || u.Host == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return true
	case "http":
		return config.IsLoopbackHost(u.Hostname())
	default:
		return false
	}
}

// checkTestKeyTarget refuses a diagnostic run that would send a key to a target
// keyMayTravelTo rejects. An empty key is always allowed.
func checkTestKeyTarget(targetURL, apiKey string) error {
	if apiKey == "" || keyMayTravelTo(targetURL) {
		return nil
	}
	return fmt.Errorf("refusing to send an API key to %s over plain HTTP: use an https:// URL or a loopback address", targetURL)
}
