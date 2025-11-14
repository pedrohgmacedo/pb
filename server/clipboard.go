package server

import (
	"fmt"
	"golang.design/x/clipboard"
	"regexp"
	"strings"
	"sync"
)

// clipboardInitError stores any error that occurred during clipboard initialization
var clipboardInitError error

// clipboarder defines the interface for clipboard operations.
type clipboarder interface {
	Copy(data []byte) error
	Paste() ([]byte, error)
}

// systemClipboard interacts with the actual system's clipboard.
type systemClipboard struct{}

func (c *systemClipboard) Copy(data []byte) error {
	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func (c *systemClipboard) Paste() ([]byte, error) {
	data := clipboard.Read(clipboard.FmtText)
	return data, nil
}

// inMemoryClipboard is used as a fallback when the system clipboard is not available.
type inMemoryClipboard struct {
	mu   sync.RWMutex
	data []byte
}

func (c *inMemoryClipboard) Copy(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	return nil
}

func (c *inMemoryClipboard) Paste() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data, nil
}

// activeClipboard holds the implementation that will be used for all clipboard operations.
var activeClipboard clipboarder

// init runs once when the package is loaded. It initializes the clipboard and sets
// the activeClipboard to the correct implementation based on availability.
func init() {
	clipboardInitError = clipboard.Init()
	if clipboardInitError != nil {
		activeClipboard = &inMemoryClipboard{}
	} else {
		activeClipboard = &systemClipboard{}
	}
}

func UseInMemoryClipboard() {
	activeClipboard = &inMemoryClipboard{}
}

// CopyToClipboard writes the given data using the active clipboard implementation.
func CopyToClipboard(data []byte) error {
	if activeClipboard == nil {
		// This should not happen in practice due to the init() function.
		return fmt.Errorf("clipboard not initialized")
	}
	return activeClipboard.Copy(data)
}

// PasteFromClipboard reads data using the active clipboard implementation.
func PasteFromClipboard() ([]byte, error) {
	if activeClipboard == nil {
		// This should not happen in practice due to the init() function.
		return nil, fmt.Errorf("clipboard not initialized")
	}
	return activeClipboard.Paste()
}

// ConvertLE is used to normalize line endings when exchanging clipboard content.
// This can be used on the client side if needed.
func ConvertLE(text, op string) string {
	switch {
	case strings.EqualFold("lf", op):
		text = strings.ReplaceAll(text, "\r\n", "\n")
		return strings.ReplaceAll(text, "\r", "\n")
	case strings.EqualFold("crlf", op):
		text = regexp.MustCompile(`\r(.)|\r$`).ReplaceAllString(text, "\r\n$1")
		text = regexp.MustCompile(`([^\r])\n|^\n`).ReplaceAllString(text, "$1\r\n")
		return text
	default:
		return text
	}
}
