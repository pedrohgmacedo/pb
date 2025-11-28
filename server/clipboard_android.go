//go:build android

package server

import (
	"context"
	"log"
	"pb/util"
)

// cliClipboard interacts with the system's clipboard using CLI tools.
type cliClipboard struct{}

func (c *cliClipboard) Copy(data []byte) error {
	return util.WriteClipboardCLI(data)
}

func (c *cliClipboard) Paste() ([]byte, error) {
	return util.ReadClipboardCLI()
}

// init runs once when the package is loaded. On Android/Termux, tries CLI tools
// first, then falls back to in-memory.
func init() {
	fallback := &inMemoryClipboard{}
	state = &clipboardState{
		fallback:        fallback,
		healthCheckDone: make(chan struct{}),
	}

	// Try CLI tools
	if util.CLIClipboardAvailable {
		state.active = &cliClipboard{}
		state.usingFallback = false
		log.Println("Using CLI clipboard tools")
		return
	}

	// Fallback to in-memory
	state.active = fallback
	state.usingFallback = true
	log.Println("CLI clipboard tools not available, using in-memory clipboard")
}

func getPrimaryClipboard() clipboarder {
	return &cliClipboard{}
}

func isClipboardResponsive() bool {
	ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		// Quick test read
		_, _ = util.ReadClipboardCLI()
		done <- true
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
