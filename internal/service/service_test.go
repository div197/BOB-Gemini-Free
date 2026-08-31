package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"text/template"
)

func TestMacOSPlistTemplate(t *testing.T) {
	cfg := ServiceConfig{
		ExePath: "/usr/local/bin/bob-gemini-free",
		Port:    9610,
		LogPath: "/tmp/daemon.log",
	}

	tmpl, err := template.New("plist").Parse(macOSPlistTemplate)
	if err != nil {
		t.Fatalf("failed to parse plist template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("failed to execute plist template: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, "com.abcsteps.bob-gemini-free") {
		t.Errorf("expected label in plist, got %s", rendered)
	}
	if !strings.Contains(rendered, "/usr/local/bin/bob-gemini-free") {
		t.Errorf("expected ExePath in plist, got %s", rendered)
	}
	if !strings.Contains(rendered, "9610") {
		t.Errorf("expected Port in plist, got %s", rendered)
	}
}

func TestLinuxSystemdTemplate(t *testing.T) {
	cfg := ServiceConfig{
		ExePath: "/usr/local/bin/bob-gemini-free",
		Port:    9610,
		LogPath: "/tmp/daemon.log",
	}

	tmpl, err := template.New("systemd").Parse(linuxSystemdTemplate)
	if err != nil {
		t.Fatalf("failed to parse systemd template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("failed to execute systemd template: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, "ExecStart=/usr/local/bin/bob-gemini-free --port 9610") {
		t.Errorf("expected ExecStart in systemd unit, got %s", rendered)
	}
}

func TestWindowsBatchTemplate(t *testing.T) {
	cfg := ServiceConfig{
		ExePath: `C:\bob-gemini-free.exe`,
		Port:    9610,
		LogPath: "",
	}

	tmpl, err := template.New("win").Parse(windowsBatchTemplate)
	if err != nil {
		t.Fatalf("failed to parse windows batch template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("failed to execute windows batch template: %v", err)
	}

	rendered := buf.String()
	if !strings.Contains(rendered, `start "" "C:\bob-gemini-free.exe" --port 9610`) {
		t.Errorf("expected batch start command, got %s", rendered)
	}
}

func TestCompatibleHealthResponseRequiresBOBHealthContract(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		gateway    string
		protocol   string
		want       bool
	}{
		{name: "valid", statusCode: http.StatusOK, gateway: serviceHealthGateway, protocol: serviceHealthProtocol, want: true},
		{name: "wrong status", statusCode: http.StatusUnauthorized, gateway: serviceHealthGateway, protocol: serviceHealthProtocol, want: false},
		{name: "missing identity", statusCode: http.StatusOK, protocol: serviceHealthProtocol, want: false},
		{name: "wrong identity", statusCode: http.StatusOK, gateway: "other-service", protocol: serviceHealthProtocol, want: false},
		{name: "wrong protocol", statusCode: http.StatusOK, gateway: serviceHealthGateway, protocol: "2", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}
			if tt.gateway != "" {
				resp.Header.Set(serviceHealthHeader, tt.gateway)
			}
			if tt.protocol != "" {
				resp.Header.Set(serviceProtocolHeader, tt.protocol)
			}
			if got := isCompatibleHealthResponse(resp); got != tt.want {
				t.Fatalf("isCompatibleHealthResponse = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusUsesDedicatedHealthEndpointAndRejectsUnrelatedProcess(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	portText := server.URL[strings.LastIndex(server.URL, ":")+1:]
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}
	var messages []string
	err = Status(port, func(format string, args ...any) {
		messages = append(messages, format)
	})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if requestedPath != serviceHealthPath {
		t.Fatalf("status probe path = %q, want %q", requestedPath, serviceHealthPath)
	}
	joined := strings.Join(messages, "\n")
	if !strings.Contains(joined, "not a compatible BOB gateway") {
		t.Fatalf("status output = %q, want unrelated process diagnostic", joined)
	}
}
