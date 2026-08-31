package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/div197/bob-gemini-free/internal/server"
)

func TestStartDesktopGatewayChoosesFallbackWhenPortBelongsToAnotherProcess(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupy port: %v", err)
	}
	defer occupied.Close()
	occupiedPort := occupied.Addr().(*net.TCPAddr).Port

	gateway, err := startDesktopGateway(occupiedPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), "test-version")
	if err != nil {
		t.Fatalf("startDesktopGateway: %v", err)
	}
	defer gateway.Shutdown(context.Background())
	if gateway.reused {
		t.Fatal("unrelated occupied port was incorrectly treated as compatible")
	}
	if gateway.Endpoint() == "http://127.0.0.1:"+strconv.Itoa(occupiedPort) {
		t.Fatalf("gateway reused occupied endpoint %s", gateway.Endpoint())
	}
	resp := waitForGatewayEndpoint(t, gateway.Endpoint())
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("fallback status = %d", resp.StatusCode)
	}
}

func TestStartDesktopGatewayReusesCompatibleGateway(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen compatible gateway: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-BOB-Gateway", "bob-gemini-free")
			w.Header().Set("X-BOB-Protocol", "1")
			w.Header().Set(server.HealthzVersionHeader, "test-version")
			w.Header().Set("X-BOB-Auth-Required", "false")
			_, _ = io.WriteString(w, `{"status":"ok"}`)
			return
		}
		http.NotFound(w, r)
	})}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())
	ready := waitForGatewayEndpoint(t, "http://127.0.0.1:"+strconv.Itoa(port)+"/healthz")
	ready.Body.Close()

	gateway, err := startDesktopGateway(port, http.NotFoundHandler(), "test-version")
	if err != nil {
		t.Fatalf("startDesktopGateway reuse: %v", err)
	}
	if !gateway.reused {
		t.Fatal("compatible gateway was not reused")
	}
	if gateway.Endpoint() != "http://127.0.0.1:"+strconv.Itoa(port) {
		t.Fatalf("reused endpoint = %q", gateway.Endpoint())
	}
	if err := gateway.Shutdown(context.Background()); err != nil {
		t.Fatalf("reused shutdown: %v", err)
	}
}

func waitForGatewayEndpoint(t *testing.T, endpoint string) *http.Response {
	t.Helper()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for {
		resp, err := client.Get(endpoint)
		if err == nil {
			return resp
		}
		lastErr = err
		if time.Now().After(deadline) {
			t.Fatalf("gateway endpoint %s did not become ready: %v", endpoint, lastErr)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestProbeCompatibleGatewayRejectsNonBOBResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"other"}`)
	}))
	defer server.Close()
	compatible, err := probeCompatibleGateway(server.URL, "test-version")
	if compatible || err == nil {
		t.Fatalf("probe result = compatible %v, err %v", compatible, err)
	}
}

func TestProbeCompatibleGatewayRejectsStatusOnlyLookalike(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"status":"ok"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	compatible, err := probeCompatibleGateway(server.URL, "test-version")
	if compatible || err == nil {
		t.Fatalf("status-only probe result = compatible %v, err %v", compatible, err)
	}
}

func TestProbeCompatibleGatewayRejectsAuthenticatedGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-BOB-Gateway", "bob-gemini-free")
			w.Header().Set("X-BOB-Protocol", "1")
			w.Header().Set(server.HealthzVersionHeader, "test-version")
			w.Header().Set("X-BOB-Auth-Required", "true")
			_, _ = io.WriteString(w, `{"status":"ok"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	compatible, err := probeCompatibleGateway(server.URL, "test-version")
	if compatible || err == nil {
		t.Fatalf("authenticated probe result = compatible %v, err %v", compatible, err)
	}
}

func TestStartDesktopGatewayDoesNotReuseDifferentVersion(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen old gateway: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	oldGateway := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-BOB-Gateway", "bob-gemini-free")
			w.Header().Set("X-BOB-Protocol", server.HealthzProtocolVersion)
			w.Header().Set(server.HealthzVersionHeader, "old-version")
			w.Header().Set("X-BOB-Auth-Required", "false")
			_, _ = io.WriteString(w, `{"status":"ok"}`)
			return
		}
		http.NotFound(w, r)
	})}
	go oldGateway.Serve(listener)
	ready := waitForGatewayEndpoint(t, "http://127.0.0.1:"+strconv.Itoa(port)+"/healthz")
	ready.Body.Close()

	gateway, err := startDesktopGateway(port, http.NotFoundHandler(), "new-version")
	if err != nil {
		t.Fatalf("startDesktopGateway version mismatch: %v", err)
	}
	defer gateway.Shutdown(context.Background())
	if gateway.reused {
		t.Fatal("older gateway version was incorrectly reused")
	}
	if gateway.Endpoint() == "http://127.0.0.1:"+strconv.Itoa(port) {
		t.Fatalf("gateway reused stale endpoint %s", gateway.Endpoint())
	}
	ready = waitForGatewayEndpoint(t, gateway.Endpoint()+"/healthz")
	ready.Body.Close()
	if err := oldGateway.Shutdown(context.Background()); err != nil {
		t.Fatalf("stop old gateway: %v", err)
	}
}
