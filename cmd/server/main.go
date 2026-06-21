package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/proxis-c2/proxis-c2/internal/common/config"
	"github.com/proxis-c2/proxis-c2/internal/common/crypto"
	"github.com/proxis-c2/proxis-c2/internal/common/protocol"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

// C2Server represents the main C2 server
type C2Server struct {
	protocol.UnimplementedC2Server
	config         *config.ServerConfig
	grpcServer     *grpc.Server
	listener       net.Listener
	agents         map[string]*AgentSession
	cryptoProvider *crypto.AESGCMProvider
}

// AgentSession represents a connected agent
type AgentSession struct {
	ID         string
	LastBeacon time.Time
	Platform   string
	Hostname   string
	Tasks      chan *protocol.Task
	OutputChan chan *protocol.TaskResult
}

// NewC2Server creates a new C2 server instance
func NewC2Server(cfg *config.ServerConfig) (*C2Server, error) {
	// Load encryption key
	key, err := os.ReadFile(cfg.Security.EncryptionKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load encryption key: %w", err)
	}

	cryptoProvider, err := crypto.NewAESGCMProvider(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crypto provider: %w", err)
	}

	return &C2Server{
		config:         cfg,
		agents:         make(map[string]*AgentSession),
		cryptoProvider: cryptoProvider,
	}, nil
}

// Start starts the C2 server
func (s *C2Server) Start() error {
	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(
		s.config.Server.TLS.CertPath,
		s.config.Server.TLS.KeyPath,
	)
	if err != nil {
		return fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA for client verification
	caCert, err := os.ReadFile(s.config.Server.TLS.CAPath)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
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
	s.grpcServer = grpc.NewServer(grpc.Creds(creds))

	// Register gRPC services
	protocol.RegisterC2Server(s.grpcServer, s)

	// Create listener
	addr := fmt.Sprintf("%s:%d", s.config.Server.ListenAddress, s.config.Server.ListenPort)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	log.Printf("Server starting on %s (version: %s, build: %s)", addr, version, buildDate)

	// Start gRPC server
	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops the C2 server
func (s *C2Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// HandleAgentHello handles new agent connections
func (s *C2Server) HandleAgentHello(ctx context.Context, req *protocol.AgentHello) (*protocol.AgentHelloResponse, error) {
	session := &AgentSession{
		ID:         req.AgentId,
		LastBeacon: time.Now(),
		Platform:   req.Platform,
		Hostname:   req.Hostname,
		Tasks:      make(chan *protocol.Task, 100),
		OutputChan: make(chan *protocol.TaskResult, 100),
	}

	s.agents[req.AgentId] = session
	log.Printf("Agent connected: %s (%s/%s)", req.AgentId, req.Hostname, req.Platform)

	return &protocol.AgentHelloResponse{
		Accepted: true,
		Message:  "Welcome to proxis-c2",
	}, nil
}

// HandleBeacon handles agent beacons
func (s *C2Server) HandleBeacon(ctx context.Context, req *protocol.Beacon) (*protocol.BeaconResponse, error) {
	session, exists := s.agents[req.AgentId]
	if !exists {
		return &protocol.BeaconResponse{
			Status:  "error",
			Message: "Unknown agent",
		}, nil
	}

	session.LastBeacon = time.Now()
	return &protocol.BeaconResponse{
		Status:  "ok",
		Message: "Beacon received",
	}, nil
}

// main is the entry point for the C2 server
func main() {
	// Load configuration
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AddConfigPath("/etc/proxis")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var cfg config.ServerConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Set defaults
	if cfg.Server.ListenAddress == "" {
		cfg.Server.ListenAddress = "0.0.0.0"
	}
	if cfg.Server.ListenPort == 0 {
		cfg.Server.ListenPort = 443
	}

	// Create and start server
	server, err := NewC2Server(&cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		server.Stop()
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}