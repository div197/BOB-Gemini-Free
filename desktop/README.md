# Legacy Tauri Desktop Wrapper

This directory is retained for historical compatibility only. It is not the
canonical desktop build and is not included by the root Makefile or local
release packaging. New desktop work must use the Wails application in
`cmd/desktop` (`make desktop`), which embeds the Go gateway and handles port
collisions and lifecycle shutdown.

The Tauri sidecar still contains the earlier fixed-port wrapper. Do not ship
it as a supported release path without a separate review.
