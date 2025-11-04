package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"
)

// GetCertsPath returns the path to certs directory for server
// Auto-detects environment and uses appropriate path:
// - If CERTS_PATH env var is set:
//   - If absolute path (starts with /), use it directly
//   - If relative path, resolve from current working directory (pwd + CERTS_PATH)
// - If running in container (/app exists), use /app/certs
// - If ../certs exists (relative path in container), use ../certs
// - Otherwise use src/server/certs for local development
func GetCertsPath() string {
	// Priority 1: Use CERTS_PATH env var if set
	if certsPath := os.Getenv("CERTS_PATH"); certsPath != "" {
		// If absolute path, use it directly
		if filepath.IsAbs(certsPath) {
			return certsPath
		}
		// If relative path, resolve from current working directory
		wd, err := os.Getwd()
		if err == nil {
			// Resolve relative path from working directory
			resolvedPath := filepath.Join(wd, certsPath)
			// Check if resolved path exists
			if _, err := os.Stat(resolvedPath); err == nil {
				return resolvedPath
			}
			// If not found, try common alternatives for relative "certs"
			if certsPath == "certs" {
				// Try src/server/certs for local development
				if altPath := filepath.Join(wd, "src", "server", "certs"); altPath != resolvedPath {
					if _, err := os.Stat(altPath); err == nil {
						return altPath
					}
				}
			}
			// Return resolved path anyway (may be created later)
			return resolvedPath
		}
		// If can't get working directory, return as-is (will be relative to binary location)
		return certsPath
	}

	// Priority 2: Check if running in container (check /app directory)
	if _, err := os.Stat("/app"); err == nil {
		// We're in a container, check if /app/certs exists
		if _, err := os.Stat("/app/certs"); err == nil {
			return "/app/certs"
		}
		// If /app/certs doesn't exist, try ../certs (relative path)
		if _, err := os.Stat("../certs"); err == nil {
			return "../certs"
		}
		// Fallback to /app/certs even if it doesn't exist yet (will be mounted)
		return "/app/certs"
	}

	// Priority 3: Check if ../certs exists (for container with relative path)
	if _, err := os.Stat("../certs"); err == nil {
		return "../certs"
	}

	// Priority 4: Local development - use src/server/certs
	// Check if src/server/certs exists
	if _, err := os.Stat("src/server/certs"); err == nil {
		return "src/server/certs"
	}

	// Fallback: default path
	return "/app/certs"
}

// LoadClientTLSCredentials loads client TLS credentials for mTLS
// serverName should be the service name: user-service, council-service, etc.
func LoadClientTLSCredentials(serverName string) (credentials.TransportCredentials, error) {
	certsPath := GetCertsPath()

	// Load client certificate and private key
	clientCert := filepath.Join(certsPath, "clients", "client.crt")
	clientKey := filepath.Join(certsPath, "clients", "client.key")

	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %v", err)
	}

	// Load CA certificate for server verification
	caCert := filepath.Join(certsPath, "clients", "ca.crt")
	caPool := x509.NewCertPool()

	ca, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	if !caPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      caPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// LoadClientTLSCredentialsInsecure loads client TLS credentials WITHOUT client certificate
// This is for one-way TLS (server authentication only)
// Only use if you want TLS encryption but not mTLS
func LoadClientTLSCredentialsInsecure(serverName string) (credentials.TransportCredentials, error) {
	certsPath := GetCertsPath()

	// Load CA certificate for server verification
	caCert := filepath.Join(certsPath, "clients", "ca.crt")
	caPool := x509.NewCertPool()

	ca, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	if !caPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create TLS configuration (no client certificate)
	tlsConfig := &tls.Config{
		RootCAs:    caPool,
		ServerName: serverName,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

