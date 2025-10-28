package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"
)

// GetCertsPath returns the path to certs directory
// For client certificates (typically in certs/clients/)
func GetCertsPath() string {
	// Try environment variable first
	if certsPath := os.Getenv("CERTS_PATH"); certsPath != "" {
		return certsPath
	}

	// Docker default
	if _, err := os.Stat("/app/certs"); err == nil {
		return "/app/certs"
	}

	// Local fallback
	return filepath.Join(os.Getenv("HOME"), "code", "heheheh_be", "certs")
}

// LoadServerTLSCredentials loads server TLS credentials for mTLS
// serviceName should be one of: user, council, thesis, academic, role, file
func LoadServerTLSCredentials(serviceName string) (credentials.TransportCredentials, error) {
	// Get cert path from env var or use default
	var basePath string

	// Try SERVICE_CERT_PATH env var first
	if certPath := os.Getenv("SERVICE_CERT_PATH"); certPath != "" {
		basePath = certPath
	} else if _, err := os.Stat("/app/service"); err == nil {
		// Docker: mounted at /app/service
		basePath = "/app/service"
	} else {
		// Local fallback - certs in src/service/{service_name}/
		basePath = filepath.Join(os.Getenv("HOME"), "code", "heheheh_be", "src", "service", serviceName)
	}

	// Load server certificate and private key
	serverCert := filepath.Join(basePath, fmt.Sprintf("%s-server.crt", serviceName))
	serverKey := filepath.Join(basePath, fmt.Sprintf("%s-server.key", serviceName))

	certificate, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	// Load CA certificate for client verification
	var caCert string

	// Try SERVICE_CA_CERT env var first
	if caCertPath := os.Getenv("SERVICE_CA_CERT"); caCertPath != "" {
		caCert = caCertPath
	} else if _, err := os.Stat("/app/service/ca.crt"); err == nil {
		// Docker: ca.crt in service folder
		caCert = "/app/service/ca.crt"
	} else {
		// Local: ca.crt in certs/ca/ folder
		caCert = filepath.Join(os.Getenv("HOME"), "code", "heheheh_be", "certs", "ca", "ca.crt")
	}

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
	// Get cert path from env var or use default
	var basePath string

	// Try SERVICE_CERT_PATH env var first
	if certPath := os.Getenv("SERVICE_CERT_PATH"); certPath != "" {
		basePath = certPath
	} else if _, err := os.Stat("/app/service"); err == nil {
		// Docker: mounted at /app/service
		basePath = "/app/service"
	} else {
		// Local fallback - certs in src/service/{service_name}/
		basePath = filepath.Join(os.Getenv("HOME"), "code", "heheheh_be", "src", "service", serviceName)
	}

	// Check service certificates
	if serviceName != "" {
		serverCert := filepath.Join(basePath, fmt.Sprintf("%s-server.crt", serviceName))
		serverKey := filepath.Join(basePath, fmt.Sprintf("%s-server.key", serviceName))

		if _, err := os.Stat(serverCert); os.IsNotExist(err) {
			return fmt.Errorf("server certificate not found: %s", serverCert)
		}
		if _, err := os.Stat(serverKey); os.IsNotExist(err) {
			return fmt.Errorf("server key not found: %s", serverKey)
		}
	}

	return nil
}
