package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx                    context.Context
	gatewayURL             string
	updateConfirmationPath string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GatewayURL returns the actual loopback endpoint selected for this desktop
// process. The frontend primarily receives it through the Wails event, while
// this binding remains available for a later explicit pairing/bootstrap flow.
func (a *App) GatewayURL() string {
	return a.gatewayURL
}
