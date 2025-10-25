package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"
)

// LoadServerTLSCredentialsOptional loads server TLS credentials with optional client verification
// This allows connections with or without client certificates
// Set requireClientCert=false for Docker/production where client certs may not be available
func LoadServerTLSCredentialsOptional(serviceName string, requireClientCert bool) (credentials.TransportCredentials, error) {
	certsPath := GetCertsPath()

	// Load server certificate and private key
	serverCert := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.crt", serviceName))
	serverKey := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.key", serviceName))

	certificate, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	// Load CA certificate for client verification (if needed)
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
	var clientAuth tls.ClientAuthType
	if requireClientCert {
		// mTLS: Client must provide valid certificate
		clientAuth = tls.RequireAndVerifyClientCert
	} else {
		// Optional mTLS: Client can connect with or without certificate
		clientAuth = tls.VerifyClientCertIfGiven
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   clientAuth,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// LoadServerTLSCredentialsInsecure loads server TLS credentials WITHOUT client verification
// WARNING: Only use in development or trusted networks
// This provides encryption but no client authentication
func LoadServerTLSCredentialsInsecure(serviceName string) (credentials.TransportCredentials, error) {
	certsPath := GetCertsPath()

	// Load server certificate and private key
	serverCert := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.crt", serviceName))
	serverKey := filepath.Join(certsPath, "services", serviceName, fmt.Sprintf("%s-server.key", serviceName))

	certificate, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	// Create TLS configuration without client verification
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.NoClientCert, // No client certificate required
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// LoadServerTLSFromEnv loads server TLS credentials based on environment variable
// Set TLS_MODE=strict for mTLS, TLS_MODE=optional for optional mTLS, TLS_MODE=none for no client auth
func LoadServerTLSFromEnv(serviceName string) (credentials.TransportCredentials, error) {
	tlsMode := os.Getenv("TLS_MODE")

	switch tlsMode {
	case "strict":
		// Require client certificate (mTLS)
		return LoadServerTLSCredentials(serviceName)
	case "optional":
		// Optional client certificate
		return LoadServerTLSCredentialsOptional(serviceName, false)
	case "none", "insecure":
		// No client certificate required
		return LoadServerTLSCredentialsInsecure(serviceName)
	default:
		// Default to optional for backward compatibility
		return LoadServerTLSCredentialsOptional(serviceName, false)
	}
}
