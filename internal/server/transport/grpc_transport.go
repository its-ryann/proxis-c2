package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCTransport implements the gRPC transport layer
type GRPCTransport struct {
	server   *grpc.Server
	listener net.Listener
}

// NewGRPCTransport creates a new gRPC transport
func NewGRPCTransport(cfg *TLSConfig) (*GRPCTransport, error) {
	// Load server certificate
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA for client verification
	caCert, err := readFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:   caPool,
		ClientAuth:  tls.RequireAndVerifyClientCert,
		MinVersion:  tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	creds := credentials.NewTLS(tlsConfig)
	server := grpc.NewServer(grpc.Creds(creds))

	return &GRPCTransport{server: server}, nil
}

// Start starts the gRPC server
func (t *GRPCTransport) Start(addr string) error {
	var err error
	t.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	return t.server.Serve(t.listener)
}

// Stop gracefully stops the gRPC server
func (t *GRPCTransport) Stop() {
	if t.server != nil {
		t.server.GracefulStop()
	}
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	CertPath string
	KeyPath  string
	CAPath   string
}

// readFile is a helper to read file contents
func readFile(path string) ([]byte, error) {
	return nil, nil // Placeholder - will be implemented with proper file reading
}