package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"
)

// GetProjectRoot returns the project root directory
func GetProjectRoot() string {
	// Try environment variable first
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}
	// Default fallback
	return "/home/thaily/code/heheheh_be"
}

// GetCertsPath returns the path to certs directory
func GetCertsPath() string {
	return filepath.Join(GetProjectRoot(), "certs")
}

// LoadServerTLSCredentials loads server TLS credentials for mTLS
// serviceName should be one of: user, council, thesis, academic, role, file
func LoadServerTLSCredentials(serviceName string) (credentials.TransportCredentials, error) {
	certsPath := GetCertsPath()

	// Load server certificate and private key
	serverCert := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.crt", serviceName))
	serverKey := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.key", serviceName))

	certificate, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	// Load CA certificate for client verification
	caCert := filepath.Join(certsPath, "ca", "ca.crt")
	caPool := x509.NewCertPool()

	ca, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	if !caPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create TLS configuration
	// ClientAuth: RequireAndVerifyClientCert means mTLS (client must provide valid cert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
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

// VerifyCertificatesExist checks if required certificate files exist
func VerifyCertificatesExist(serviceName string) error {
	certsPath := GetCertsPath()

	// Check CA
	caCert := filepath.Join(certsPath, "ca", "ca.crt")
	if _, err := os.Stat(caCert); os.IsNotExist(err) {
		return fmt.Errorf("CA certificate not found: %s\nRun: cd %s && ./generate-certs.sh", caCert, certsPath)
	}

	// Check service certificates
	if serviceName != "" {
		serverCert := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.crt", serviceName))
		serverKey := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.key", serviceName))

		if _, err := os.Stat(serverCert); os.IsNotExist(err) {
			return fmt.Errorf("server certificate not found: %s\nRun: cd %s && ./generate-certs.sh", serverCert, certsPath)
		}
		if _, err := os.Stat(serverKey); os.IsNotExist(err) {
			return fmt.Errorf("server key not found: %s\nRun: cd %s && ./generate-certs.sh", serverKey, certsPath)
		}
	}

	return nil
}
