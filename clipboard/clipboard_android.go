//go:build android

package clipboard

import (
	"context"
)

// cliClipboard interacts with the system's clipboard using CLI tools.
type cliClipboard struct{}

func (c *cliClipboard) Copy(data []byte) error {
	return WriteClipboardCLI(data)
}

func (c *cliClipboard) Paste() ([]byte, error) {
	return ReadClipboardCLI()
}

// initPlatformClipboard tries CLI tools first, then falls back to in-memory.
func initPlatformClipboard(fallback *inMemoryClipboard) error {
	// Try CLI tools
	if CLIClipboardAvailable {
		state.active = &cliClipboard{}
		state.usingFallback = false
		logf("Using CLI clipboard tools")
		return nil
	}

	// Fallback to in-memory
	state.active = fallback
	state.usingFallback = true
	logf("CLI clipboard tools not available, using in-memory clipboard")
	return nil
}

func getPrimaryClipboard() clipboarder {
	return &cliClipboard{}
}

func getCLIClipboard() clipboarder {
	return &cliClipboard{}
}

func isClipboardResponsive() bool {
	ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		// Quick test read
		_, _ = ReadClipboardCLI()
		done <- true
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
