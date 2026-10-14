package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// ErrEmptyListenHost is returned when the configured host is blank. An empty
// host makes net.JoinHostPort produce ":port", which binds every interface.
var ErrEmptyListenHost = errors.New("host must not be empty; use 127.0.0.1 for local-only access")

// CheckListenPolicy decides whether the gateway may bind the configured host.
// A loopback host is always allowed. Any other host (0.0.0.0, a LAN address,
// a hostname) makes the gateway reachable from other machines, so it requires
// at least one API key. Callers run this before binding a listener.
func CheckListenPolicy(cfg Config) error {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return ErrEmptyListenHost
	}
	if IsLoopbackHost(host) {
		return nil
	}
	if len(cfg.APIKeys) == 0 {
		return fmt.Errorf("refusing to listen on %q without api_keys: set api_keys in config.json or BOB_GEMINI_FREE_API_KEYS, or bind to 127.0.0.1", host)
	}
	return nil
}

// IsLoopbackHost reports whether host names only this machine: "localhost" or
// a loopback IP literal. Any other name is treated as reachable from outside.
func IsLoopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
