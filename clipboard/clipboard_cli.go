package clipboard

import (
	"errors"
	"os"
	"os/exec"
)

const (
	cliXsel               = "xsel"
	cliXclip              = "xclip"
	cliWlcopy             = "wl-copy"
	cliWlpaste            = "wl-paste"
	cliTermuxClipboardGet = "termux-clipboard-get"
	cliTermuxClipboardSet = "termux-clipboard-set"
)

var (
	// CLIClipboardAvailable indicates whether clipboard CLI tools are available
	CLIClipboardAvailable = false

	pasteCmdArgs []string
	copyCmdArgs  []string

	xselPasteArgs = []string{cliXsel, "--output", "--clipboard"}
	xselCopyArgs  = []string{cliXsel, "--input", "--clipboard"}

	xclipPasteArgs = []string{cliXclip, "-out", "-selection", "clipboard"}
	xclipCopyArgs  = []string{cliXclip, "-in", "-selection", "clipboard"}

	wlpasteArgs = []string{cliWlpaste, "--no-newline"}
	wlcopyArgs  = []string{cliWlcopy}

	termuxPasteArgs = []string{cliTermuxClipboardGet}
	termuxCopyArgs  = []string{cliTermuxClipboardSet}

	clipboardUnavailableErr = errors.New("no clipboard utilities available: install xsel, xclip, wl-clipboard, or enable Termux:API")
)

// initCLIClipboard detects available clipboard CLI tools
func initCLIClipboard() {
	// Try Wayland first
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if hasCommand(cliWlcopy) && hasCommand(cliWlpaste) {
			pasteCmdArgs = wlpasteArgs
			copyCmdArgs = wlcopyArgs
			CLIClipboardAvailable = true
			return
		}
	}

	// Try xclip
	if hasCommand(cliXclip) {
		pasteCmdArgs = xclipPasteArgs
		copyCmdArgs = xclipCopyArgs
		CLIClipboardAvailable = true
		return
	}

	// Try xsel
	if hasCommand(cliXsel) {
		pasteCmdArgs = xselPasteArgs
		copyCmdArgs = xselCopyArgs
		CLIClipboardAvailable = true
		return
	}

	// Try Termux
	if hasCommand(cliTermuxClipboardSet) && hasCommand(cliTermuxClipboardGet) {
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
