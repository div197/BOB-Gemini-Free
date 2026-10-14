package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/div197/bob-gemini-free/internal/config"
)

// keyMayTravelTo reports whether a gateway API key may be sent to targetURL
// without crossing the network in plain text: an https:// target, or an
// http:// target on a loopback address.
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

// resolveTestKey chooses the key for a diagnostic run and returns an operator
// note when one applies. An explicit --test-key is always used, with a warning
// if it will go over plain HTTP. The key from config.json is a convenience
// fallback, so it is only used for https:// or loopback targets.
func resolveTestKey(explicitKey string, configuredKeys []string, targetURL string) (string, string) {
	plain := !keyMayTravelTo(targetURL)
	if explicitKey != "" {
		if plain {
			return explicitKey, fmt.Sprintf("warning: --test-key is sent to %s over plain HTTP", targetURL)
		}
		return explicitKey, ""
	}
	if len(configuredKeys) == 0 {
		return "", ""
	}
	if plain {
		return "", fmt.Sprintf("note: not sending the configured api_keys to %s over plain HTTP; pass --test-key to send one", targetURL)
	}
	return configuredKeys[0], ""
}
