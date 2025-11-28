package clipboard

import (
	"errors"
	"os"
	"os/exec"
)

const (
	xsel               = "xsel"
	xclip              = "xclip"
	wlcopy             = "wl-copy"
	wlpaste            = "wl-paste"
	termuxClipboardGet = "termux-clipboard-get"
	termuxClipboardSet = "termux-clipboard-set"
)

var (
	// CLIClipboardAvailable indicates whether clipboard CLI tools are available
	CLIClipboardAvailable = false

	pasteCmdArgs []string
	copyCmdArgs  []string

	xselPasteArgs = []string{xsel, "--output", "--clipboard"}
	xselCopyArgs  = []string{xsel, "--input", "--clipboard"}

	xclipPasteArgs = []string{xclip, "-out", "-selection", "clipboard"}
	xclipCopyArgs  = []string{xclip, "-in", "-selection", "clipboard"}

	wlpasteArgs = []string{wlpaste, "--no-newline"}
	wlcopyArgs  = []string{wlcopy}

	termuxPasteArgs = []string{termuxClipboardGet}
	termuxCopyArgs  = []string{termuxClipboardSet}

	clipboardUnavailableErr = errors.New("no clipboard utilities available: install xsel, xclip, wl-clipboard, or enable Termux:API")
)

// initCLIClipboard detects available clipboard CLI tools
func initCLIClipboard() {
	// Try Wayland first
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if hasCommand(wlcopy) && hasCommand(wlpaste) {
			pasteCmdArgs = wlpasteArgs
			copyCmdArgs = wlcopyArgs
			CLIClipboardAvailable = true
			return
		}
	}

	// Try xclip
	if hasCommand(xclip) {
		pasteCmdArgs = xclipPasteArgs
		copyCmdArgs = xclipCopyArgs
		CLIClipboardAvailable = true
		return
	}

	// Try xsel
	if hasCommand(xsel) {
		pasteCmdArgs = xselPasteArgs
		copyCmdArgs = xselCopyArgs
		CLIClipboardAvailable = true
		return
	}

	// Try Termux
	if hasCommand(termuxClipboardSet) && hasCommand(termuxClipboardGet) {
		pasteCmdArgs = termuxPasteArgs
		copyCmdArgs = termuxCopyArgs
		CLIClipboardAvailable = true
		return
	}
}

// hasCommand checks if a command is available in the system PATH
func hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ReadClipboardCLI reads data from the system clipboard using external CLI tools
func ReadClipboardCLI() ([]byte, error) {
	if !CLIClipboardAvailable {
		return nil, clipboardUnavailableErr
	}

	cmd := exec.Command(pasteCmdArgs[0], pasteCmdArgs[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WriteClipboardCLI writes data to the system clipboard using external CLI tools
func WriteClipboardCLI(data []byte) error {
	if !CLIClipboardAvailable {
		return clipboardUnavailableErr
	}

	cmd := exec.Command(copyCmdArgs[0], copyCmdArgs[1:]...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := in.Write(data); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func init() {
	initCLIClipboard()
}
