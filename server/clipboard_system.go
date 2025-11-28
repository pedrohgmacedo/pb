//go:build !android

package server

import (
	"context"
	"golang.design/x/clipboard"
	"log"
	"pb/util"
)

// systemClipboard interacts with the actual system's clipboard using golang.design.
type systemClipboard struct{}

func (c *systemClipboard) Copy(data []byte) error {
	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func (c *systemClipboard) Paste() ([]byte, error) {
	data := clipboard.Read(clipboard.FmtText)
	return data, nil
}

// cliClipboard interacts with the system's clipboard using CLI tools.
type cliClipboard struct{}

func (c *cliClipboard) Copy(data []byte) error {
	return util.WriteClipboardCLI(data)
}

func (c *cliClipboard) Paste() ([]byte, error) {
	return util.ReadClipboardCLI()
}

// init runs once when the package is loaded. On desktop systems, tries golang.design
// first, then CLI tools, then falls back to in-memory.
func init() {
	fallback := &inMemoryClipboard{}
	state = &clipboardState{
		fallback:        fallback,
		healthCheckDone: make(chan struct{}),
	}

	// Try golang.design first
	err := clipboard.Init()
	if err == nil {
		state.active = &systemClipboard{}
		state.usingFallback = false
		log.Println("Using system clipboard (golang.design)")
		return
	}

	log.Printf("System clipboard (golang.design) failed: %v", err)

	// Fall back to CLI tools if available
	if util.CLIClipboardAvailable {
		state.active = &cliClipboard{}
		state.usingFallback = false
		log.Println("Falling back to CLI clipboard tools")
		return
	}

	// Last resort: in-memory clipboard
	state.active = fallback
	state.usingFallback = true
	log.Println("No clipboard utilities available, using in-memory clipboard")
}

func getPrimaryClipboard() clipboarder {
	return &systemClipboard{}
}

func isClipboardResponsive() bool {
	ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		// Quick test read
		_ = clipboard.Read(clipboard.FmtText)
		done <- true
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
