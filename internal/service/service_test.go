package service

import (
	"bytes"
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
