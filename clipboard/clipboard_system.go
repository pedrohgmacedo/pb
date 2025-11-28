//go:build !android

package clipboard

import (
	"context"
	xclip "golang.design/x/clipboard"
	"pb/util"
)

// systemClipboard interacts with the actual system's clipboard using golang.design.
type systemClipboard struct{}

func (c *systemClipboard) Copy(data []byte) error {
	xclip.Write(xclip.FmtText, data)
	return nil
}

func (c *systemClipboard) Paste() ([]byte, error) {
	data := xclip.Read(xclip.FmtText)
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

// initPlatformClipboard tries golang.design first, then CLI tools, then falls back to in-memory.
func initPlatformClipboard(fallback *inMemoryClipboard) error {
	// Try golang.design first
	err := xclip.Init()
	if err == nil {
		state.active = &systemClipboard{}
		state.usingFallback = false
		logf("Using system clipboard (golang.design)")
		return nil
	}

	logf("System clipboard (golang.design) failed: %v", err)

	// Fall back to CLI tools if available
	if util.CLIClipboardAvailable {
		state.active = &cliClipboard{}
		state.usingFallback = false
		logf("Falling back to CLI clipboard tools")
		return nil
	}

	// Last resort: in-memory clipboard
	state.active = fallback
	state.usingFallback = true
	logf("No clipboard utilities available, using in-memory clipboard")
	return nil
}

func getPrimaryClipboard() clipboarder {
	return &systemClipboard{}
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
		_ = xclip.Read(xclip.FmtText)
		done <- true
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
