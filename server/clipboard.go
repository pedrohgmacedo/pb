package server

import (
	"fmt"
	"github.com/atotto/clipboard"
	"regexp"
	"strings"
	"sync"
)

// ClipboardUnsupported re-exports the value from the underlying library
var ClipboardUnsupported = clipboard.Unsupported

// clipboarder defines the interface for clipboard operations.
type clipboarder interface {
	Copy(text string) error
	Paste() (string, error)
}

// systemClipboard interacts with the actual system's clipboard.
type systemClipboard struct{}

func (c *systemClipboard) Copy(text string) error {
	return clipboard.WriteAll(text)
}

func (c *systemClipboard) Paste() (string, error) {
	return clipboard.ReadAll()
}

// inMemoryClipboard is used as a fallback when the system clipboard is not available.
type inMemoryClipboard struct {
	mu   sync.RWMutex
	text string
}

func (c *inMemoryClipboard) Copy(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = text
	return nil
}

func (c *inMemoryClipboard) Paste() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.text, nil
}

// activeClipboard holds the implementation that will be used for all clipboard operations.
var activeClipboard clipboarder

// init runs once when the package is loaded. It checks for system clipboard
// support and sets the activeClipboard to the correct implementation.
func init() {
	if ClipboardUnsupported {
		activeClipboard = &inMemoryClipboard{}
	} else {
		activeClipboard = &systemClipboard{}
	}
}

func UseInMemoryClipboard() {
	activeClipboard = &inMemoryClipboard{}
}

// CopyToClipboard writes the given text using the active clipboard implementation.
func CopyToClipboard(text string) error {
	if activeClipboard == nil {
		// This should not happen in practice due to the init() function.
		return fmt.Errorf("clipboard not initialized")
	}
	return activeClipboard.Copy(text)
}

// PasteFromClipboard reads text using the active clipboard implementation.
func PasteFromClipboard() (string, error) {
	if activeClipboard == nil {
		// This should not happen in practice due to the init() function.
		return "", fmt.Errorf("clipboard not initialized")
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
