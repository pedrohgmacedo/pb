package commands

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"pb/util"
)

// findPrivateKey automatically detects a private key file based on a specific priority.
func findPrivateKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Priority 1: program-specific key
	programKeyPath := filepath.Join(home, ".config", util.ProgramName, "id_ed25519")
	if _, err := os.Stat(programKeyPath); err == nil {
		return programKeyPath, nil
	}

	// Priority 2: Standard SSH keys
	sshDir := filepath.Join(home, ".ssh")
	defaultKeys := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
	for _, keyFile := range defaultKeys {
		path := filepath.Join(sshDir, keyFile)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Priority 3: Fail with a helpful message
	return "", fmt.Errorf("no private key found. Please run '%s key-gen' to create a new key, or specify one with the --key flag", util.ProgramName)
}

// getSigner finds and parses a private key, returning an ssh.Signer.
// It respects the --key flag and the prioritized search path.
func getSigner() (ssh.Signer, error) {
	// If --key flag was not used, find a key automatically.
	var pathToKey string
	if keyPath != "" {
		pathToKey = keyPath
	} else {
		var err error
		pathToKey, err = findPrivateKey()
		if err != nil {
			return nil, err
		}
	}

	privateKeyBytes, err := os.ReadFile(pathToKey)
	if err != nil {
		return nil, fmt.Errorf("could not read private key at %s: %w", pathToKey, err)
	}

	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %w", err)
	}
	return signer, nil
}

// doHTTPSRequest handles the client-side logic for creating and sending a signed HTTPS request.
func doHTTPSRequest(method, url, data string) (string, error) {
	signer, err := getSigner()
	if err != nil {
		return "", err
	}

	payloadHash := sha256.Sum256([]byte(data))
	signature, err := signer.Sign(rand.Reader, payloadHash[:])
	if err != nil {
		return "", fmt.Errorf("could not sign payload: %w", err)
	}

	// This client is insecure and trusts any server certificate.
	// This is acceptable because we are authenticating the server via our SSH key model.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}

	req.Header.Set(util.HeaderFingerprint, ssh.FingerprintSHA256(signer.PublicKey()))
	// Marshal the entire signature object, not just the blob
	signatureBytes := ssh.Marshal(signature)
	req.Header.Set(util.HeaderSignature, base64.StdEncoding.EncodeToString(signatureBytes))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned non-200 status: %d\n%s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
