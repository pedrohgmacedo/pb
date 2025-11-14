// Package server handles requests from the client
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"pb/util"
	"time"
)

// Serve starts the HTTPS server.
func Serve(ctx context.Context, port int, le string, fallback bool) error {
	// If --fallback flag is set, use in-memory clipboard from the start
	if fallback {
		UseInMemoryClipboard()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	authorizedKeys, err := loadAuthorizedKeys(filepath.Join(home, ".config", util.ProgramName, "authorized_keys"))
	if err != nil {
		return fmt.Errorf("could not load authorized keys: %w", err)
	}

	certPath := filepath.Join(home, ".config", util.ProgramName, "cert.pem")
	keyPath := filepath.Join(home, ".config", util.ProgramName, "key.pem")

	if err := generateSelfSignedCert(certPath, keyPath); err != nil {
		return fmt.Errorf("could not generate self-signed certificate: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(util.RequestCopy, copyHandler)
	mux.HandleFunc(util.RequestPaste, pasteHandler)
	mux.HandleFunc(util.RequestOpen, openHandler)
	mux.HandleFunc(util.RequestQuit, quitHandler)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: authMiddleware(mux, authorizedKeys),
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	log.Printf("%s server listening on %s", util.ProgramName, addr)
	return server.ListenAndServeTLS(certPath, keyPath)
}

func authMiddleware(next http.Handler, authorizedKeys map[string]ssh.PublicKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyFingerprint := r.Header.Get(util.HeaderFingerprint)
		signatureB64 := r.Header.Get(util.HeaderSignature)

		if keyFingerprint == "" || signatureB64 == "" {
			http.Error(w, "Missing authentication headers", http.StatusUnauthorized)
			return
		}

		pubKey, ok := authorizedKeys[keyFingerprint]
		if !ok {
			http.Error(w, "Unknown public key", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// Because ReadAll consumes the body, we need to put it back for the actual handler.
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		hash := sha256.Sum256(body)

		signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
		if err != nil {
			http.Error(w, "Invalid signature encoding", http.StatusBadRequest)
			return
		}

		sshSig := &ssh.Signature{}
		if err := ssh.Unmarshal(signatureBytes, sshSig); err != nil {
			http.Error(w, "Invalid SSH signature format", http.StatusBadRequest)
			return
		}

		if err := pubKey.Verify(hash[:], sshSig); err != nil {
			http.Error(w, "Signature verification failed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func copyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if err := CopyToClipboard(body); err != nil {
		http.Error(w, "Failed to write to clipboard", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Copy request successfully handled")
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	content, err := PasteFromClipboard()
	if err != nil {
		http.Error(w, "Failed to read from clipboard", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(content); err != nil {
		log.Printf("Failed to write response: %v", err)
	} else {
		log.Println("Paste request successfully handled")
	}
}

func openHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	urlToOpen := string(body)
	log.Printf("Open request received: '%s'", urlToOpen)

	if err := open.Run(urlToOpen); err != nil {
		http.Error(w, "Failed to open URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Open request successfully handled")
}

func quitHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Shutting down server...")
	w.WriteHeader(http.StatusOK)
	os.Exit(0)
}

func loadAuthorizedKeys(path string) (map[string]ssh.PublicKey, error) {
	authorizedKeys := make(map[string]ssh.PublicKey)

	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("authorized_keys file not found at %s. Server starting with no authorized keys.", path)
			log.Printf("Use '%s key-add' to authorize a client.\n", util.ProgramName)
			return authorizedKeys, nil // Return empty map, not an error
		}
		return nil, err // Return error for other file system issues
	}

	for len(bytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(bytes)
		if err != nil {
			// Log the error but continue, in case of a malformed line
			log.Printf("Could not parse authorized key: %v", err)
			bytes = rest
			continue
		}

		fingerprint := ssh.FingerprintSHA256(pubKey)
		authorizedKeys[fingerprint] = pubKey
		bytes = rest
	}

	log.Printf("Loaded %d authorized keys from %s", len(authorizedKeys), path)
	return authorizedKeys, nil
}

func generateSelfSignedCert(certPath, keyPath string) error {
	if _, err := os.Stat(certPath); err == nil {
		// Certificate already exists
		return nil
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(certPath), 0700); err != nil {
		return fmt.Errorf("could not create cert directory: %w", err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{util.ProgramName},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10), // 10 years

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}
