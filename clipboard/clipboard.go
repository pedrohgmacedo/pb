package clipboard

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	loggingEnabled = false
	state          *clipboardState
)

const (
	clipboardTimeout    = 2 * time.Second
	healthCheckInterval = 5 * time.Second
)

// clipboarder defines the interface for clipboard operations.
type clipboarder interface {
	Copy(data []byte) error
	Paste() ([]byte, error)
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

// clipboardState tracks which clipboard implementation is active
type clipboardState struct {
	mu              sync.RWMutex
	active          clipboarder
	fallback        *inMemoryClipboard
	usingFallback   bool
	healthCheckDone chan struct{} // signals health check to stop
}

// EnableLogging turns on logging for clipboard operations
func EnableLogging() {
	loggingEnabled = true
}

// logf conditionally logs based on loggingEnabled flag
func logf(format string, args ...interface{}) {
	if loggingEnabled {
		log.Printf(format, args...)
	}
}

// Init initializes the clipboard system. Call EnableLogging() before this if you want logging.
func Init() error {
	if state != nil {
		return nil // already initialized
	}

	fallback := &inMemoryClipboard{}
	state = &clipboardState{
		fallback:        fallback,
		healthCheckDone: make(chan struct{}),
	}

	return initPlatformClipboard(fallback)
}

func UseInMemoryClipboard() {
	if state == nil {
		return
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	state.active = state.fallback
	state.usingFallback = true
	logf("Switched to in-memory clipboard (manual flag)")
}

// UseCliClipboard switches to CLI-based clipboard if available
func UseCliClipboard() error {
	if state == nil {
		return fmt.Errorf("clipboard not initialized")
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	state.active = getCLIClipboard()
	state.usingFallback = false
	logf("Switched to CLI clipboard tools (manual flag)")
	return nil
}

// getActiveClipboard returns the currently active clipboard implementation
func getActiveClipboard() clipboarder {
	if state == nil {
		return nil
	}
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.active
}

// isUsingFallback returns whether we're currently on fallback
func isUsingFallback() bool {
	if state == nil {
		return true
	}
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.usingFallback
}

// switchToFallback switches to the fallback clipboard and starts health check
func switchToFallback() {
	if state == nil {
		return
	}
	state.mu.Lock()
	wasUsingFallback := state.usingFallback
	state.active = state.fallback
	state.usingFallback = true
	state.mu.Unlock()

	if !wasUsingFallback {
		logf("System clipboard unresponsive, switched to in-memory fallback (health check polling every 5s)")
		go startHealthCheck()
	}
}

// switchToSystem switches back to the system clipboard and stops health check
func switchToSystem() {
	if state == nil {
		return
	}
	state.mu.Lock()
	state.active = getPrimaryClipboard()
	state.usingFallback = false
	state.mu.Unlock()

	logf("System clipboard recovered, switched back from fallback")
	// Signal health check to stop
	select {
	case state.healthCheckDone <- struct{}{}:
	default:
	}
}

// Copy writes the given data with timeout and auto-switching
func Copy(data []byte) error {
	active := getActiveClipboard()
	if active == nil {
		return fmt.Errorf("clipboard not initialized")
	}

	// For fallback, no timeout needed (it's local and fast)
	if isUsingFallback() {
		return active.Copy(data)
	}

	// For system clipboard, use timeout
	ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- active.Copy(data)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		switchToFallback()
		// Retry with fallback
		return state.fallback.Copy(data)
	}
}

// Paste reads data with timeout and auto-switching
func Paste() ([]byte, error) {
	active := getActiveClipboard()
	if active == nil {
		return nil, fmt.Errorf("clipboard not initialized")
	}

	// For fallback, no timeout needed (it's local and fast)
	if isUsingFallback() {
		return active.Paste()
	}

	// For system clipboard, use timeout
	ctx, cancel := context.WithTimeout(context.Background(), clipboardTimeout)
	defer cancel()

	done := make(chan []byte, 1)
	doneErr := make(chan error, 1)
	go func() {
		data, err := active.Paste()
		if err != nil {
			doneErr <- err
		} else {
			done <- data
		}
	}()

	select {
	case data := <-done:
		return data, nil
	case err := <-doneErr:
		return nil, err
	case <-ctx.Done():
		switchToFallback()
		// Retry with fallback
		return state.fallback.Paste()
	}
}

// startHealthCheck polls the clipboard every 5s to detect recovery
func startHealthCheck() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-state.healthCheckDone:
			return
		case <-ticker.C:
			if isClipboardResponsive() {
				switchToSystem()
				return
			}
		}
	}
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

// Platform-specific functions to be implemented in clipboard_system.go or clipboard_android.go
func initPlatformClipboard(fallback *inMemoryClipboard) error
func getPrimaryClipboard() clipboarder
func getCLIClipboard() clipboarder
func isClipboardResponsive() bool
