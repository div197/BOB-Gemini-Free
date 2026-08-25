package main

import "testing"

func TestDesktopOptionsKeepNativeWindowMaximizable(t *testing.T) {
	opts := desktopOptions(NewApp(), nil, nil)
	if opts.DisableResize {
		t.Fatal("desktop window unexpectedly disables resizing")
	}
	if opts.Mac == nil {
		t.Fatal("macOS options are required to keep the native zoom button enabled")
	}
	if opts.Mac.DisableZoom {
		t.Fatal("desktop window unexpectedly disables the native macOS zoom button")
	}
}
