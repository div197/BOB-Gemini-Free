package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/div197/bob-gemini-free/internal/server"
)

type desktopGateway struct {
	endpoint string
	server   *http.Server
	reused   bool
}

func startDesktopGateway(cfgPort int, handler http.Handler, expectedVersion string) (*desktopGateway, error) {
	if cfgPort <= 0 {
		cfgPort = 9610
	}
	expectedVersion = strings.TrimSpace(expectedVersion)
	if expectedVersion == "" {
		expectedVersion = "dev"
	}
	if handler == nil {
		handler = http.NotFoundHandler()
	}

	const host = "127.0.0.1"
	requestedAddr := net.JoinHostPort(host, strconv.Itoa(cfgPort))
	listener, err := net.Listen("tcp", requestedAddr)
	if err != nil {
		requestedEndpoint := "http://" + requestedAddr
		if compatible, probeErr := probeCompatibleGateway(requestedEndpoint, expectedVersion); compatible {
			return &desktopGateway{endpoint: requestedEndpoint, reused: true}, nil
		} else if probeErr != nil {
			fmt.Printf("Desktop gateway port %s is occupied and is not a compatible BOB gateway: %v\n", requestedAddr, probeErr)
		}

		listener, err = net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			return nil, fmt.Errorf("requested gateway port %d is unavailable and no fallback port could be bound: %w", cfgPort, err)
		}
	}

	port := listener.Addr().(*net.TCPAddr).Port
	endpoint := "http://" + net.JoinHostPort(host, strconv.Itoa(port))
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	gateway := &desktopGateway{endpoint: endpoint, server: srv}
	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			fmt.Printf("Desktop gateway stopped unexpectedly: %v\n", serveErr)
		}
	}()
	return gateway, nil
}

func (g *desktopGateway) Endpoint() string {
	if g == nil {
		return ""
	}
	return g.endpoint
}

func (g *desktopGateway) Shutdown(ctx context.Context) error {
	if g == nil || g.reused || g.server == nil {
		return nil
	}
	return g.server.Shutdown(ctx)
}

func probeCompatibleGateway(endpoint, expectedVersion string) (bool, error) {
	expectedVersion = strings.TrimSpace(expectedVersion)
	if expectedVersion == "" {
		expectedVersion = "dev"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/healthz", nil)
	if err != nil {
		return false, err
	}
	resp, err := (&http.Client{Timeout: 750 * time.Millisecond}).Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("healthz returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<10)).Decode(&payload); err != nil {
		return false, fmt.Errorf("invalid healthz JSON: %w", err)
	}
	if payload.Status != "ok" {
		return false, fmt.Errorf("healthz status %q", payload.Status)
	}
	if resp.Header.Get("X-BOB-Gateway") != "bob-gemini-free" {
		return false, fmt.Errorf("healthz did not identify BOB Gemini Free")
	}
	if resp.Header.Get("X-BOB-Protocol") != server.HealthzProtocolVersion {
		return false, fmt.Errorf("unsupported BOB health protocol %q", resp.Header.Get("X-BOB-Protocol"))
	}
	if got := strings.TrimSpace(resp.Header.Get(server.HealthzVersionHeader)); got != expectedVersion {
		return false, fmt.Errorf("BOB gateway version %q does not match expected %q", got, expectedVersion)
	}
	if resp.Header.Get("X-BOB-Auth-Required") != "false" {
		return false, fmt.Errorf("existing gateway requires API-key authentication")
	}
	return true, nil
}
