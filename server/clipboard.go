package server

import (
	"context"
	"fmt"
	"golang.design/x/clipboard"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
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

// clipboardState tracks whether we're using system or fallback clipboard
type clipboardState struct {
	mu              sync.RWMutex
	active          clipboarder
	fallback        *inMemoryClipboard
	usingFallback   bool
	healthCheckDone chan struct{} // signals health check to stop
}

var state *clipboardState

const (
	clipboardTimeout      = 2 * time.Second
	healthCheckInterval   = 5 * time.Second
)

// init runs once when the package is loaded. It initializes the clipboard and sets
// the activeClipboard to the correct implementation based on availability.
func init() {
	fallback := &inMemoryClipboard{}
	state = &clipboardState{
		fallback:        fallback,
		healthCheckDone: make(chan struct{}),
	}

	clipboardInitError = clipboard.Init()
	if clipboardInitError != nil {
		state.active = fallback
		state.usingFallback = true
		log.Println("System clipboard not available at startup, using in-memory clipboard")
	} else {
		state.active = &systemClipboard{}
		state.usingFallback = false
	}
}

func UseInMemoryClipboard() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.active = state.fallback
	state.usingFallback = true
	log.Println("Switched to in-memory clipboard (manual flag)")
}

// getActiveClipboard returns the currently active clipboard implementation
func getActiveClipboard() clipboarder {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.active
}

// isUsingFallback returns whether we're currently on fallback
func isUsingFallback() bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.usingFallback
}

// switchToFallback switches to the fallback clipboard and starts health check
func switchToFallback() {
	state.mu.Lock()
	wasUsingFallback := state.usingFallback
	state.active = state.fallback
	state.usingFallback = true
	state.mu.Unlock()

	if !wasUsingFallback {
		log.Println("System clipboard unresponsive, switched to in-memory fallback (health check polling every 5s)")
		go startHealthCheck()
	}
}

// switchToSystem switches back to the system clipboard and stops health check
func switchToSystem() {
	state.mu.Lock()
	state.active = &systemClipboard{}
	state.usingFallback = false
	state.mu.Unlock()

	log.Println("System clipboard recovered, switched back from fallback")
	// Signal health check to stop
	select {
	case state.healthCheckDone <- struct{}{}:
	default:
	}
}

// CopyToClipboard writes the given data with timeout and auto-switching
func CopyToClipboard(data []byte) error {
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

// PasteFromClipboard reads data with timeout and auto-switching
func PasteFromClipboard() ([]byte, error) {
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

// startHealthCheck polls the system clipboard every 5s to detect recovery
func startHealthCheck() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-state.healthCheckDone:
			return
		case <-ticker.C:
			if isSystemClipboardResponsive() {
				switchToSystem()
				return
			}
		}
	}
}

// isSystemClipboardResponsive tests if the system clipboard is accessible
func isSystemClipboardResponsive() bool {
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
