package util

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
)

// GenerateSSHKeys creates a new ed25519 SSH key pair in the specified directory.
func GenerateSSHKeys(keyDir string) error {
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("cannot create keys directory %s: %w", keyDir, err)
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("cannot generate ed25519 key: %w", err)
	}

	// Encode private key to PEM format using the standard library
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("could not marshal private key: %w", err)
	}
	privBlock := pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: pkcs8Key,
	}
	privatePEM := pem.EncodeToMemory(&privBlock)
	err = os.WriteFile(filepath.Join(keyDir, "id_ed25519"), privatePEM, 0600)
	if err != nil {
		return fmt.Errorf("unable to save private key: %w", err)
	}

	// Public key
	publicKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("unable to generate public key: %w", err)
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	err = os.WriteFile(filepath.Join(keyDir, "id_ed25519.pub"), pubKeyBytes, 0644)
	if err != nil {
		return fmt.Errorf("unable to save public key: %w", err)
	}

	return nil
}
