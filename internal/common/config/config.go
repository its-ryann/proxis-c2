package config

import (
	"time"
)

// ServerConfig contains the server configuration
type ServerConfig struct {
	Server   ServerSettings   `mapstructure:"server"`
	Database DatabaseConfig   `mapstructure:"database"`
	Logging  LoggingConfig    `mapstructure:"logging"`
	Security SecurityConfig  `mapstructure:"security"`
}

// ServerSettings contains server network settings
type ServerSettings struct {
	ListenAddress     string        `mapstructure:"listen_address"`
	ListenPort        int           `mapstructure:"listen_port"`
	TLS               TLSConfig     `mapstructure:"tls"`
	MaxConnections    int           `mapstructure:"max_connections"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	PoolSize int    `mapstructure:"pool_size"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	EncryptionKeyPath string `mapstructure:"encryption_key_path"`
	HMACSecretPath    string `mapstructure:"hmac_secret_path"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	CertPath         string `mapstructure:"cert_path"`
	KeyPath          string `mapstructure:"key_path"`
	CAPath           string `mapstructure:"ca_path"`
	RequireClientCert bool   `mapstructure:"require_client_cert"`
}

// AgentConfig contains the agent configuration
type AgentConfig struct {
	C2        C2Config        `mapstructure:"c2"`
	Beacon    BeaconConfig    `mapstructure:"beacon"`
	Evasion   EvasionConfig   `mapstructure:"evasion"`
	KillSwitch KillSwitchConfig `mapstructure:"kill_switch"`
}

// C2Config contains C2 server connection settings
type C2Config struct {
	ServerAddress    string `mapstructure:"server_address"`
	ServerPort       int    `mapstructure:"server_port"`
	UseTLS           bool   `mapstructure:"use_tls"`
	CACertPath       string `mapstructure:"ca_cert_path"`
	ClientCertPath   string `mapstructure:"client_cert_path"`
	ClientKeyPath    string `mapstructure:"client_key_path"`
}

// BeaconConfig contains beacon interval settings
type BeaconConfig struct {
	BaseInterval  time.Duration `mapstructure:"base_interval"`
	JitterMin     float64       `mapstructure:"jitter_min"`
	JitterMax     float64       `mapstructure:"jitter_max"`
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryBackoff  time.Duration `mapstructure:"retry_backoff"`
}

// EvasionConfig contains evasion detection settings
type EvasionConfig struct {
	SandboxDetection  bool          `mapstructure:"sandbox_detection"`
	DebuggerDetection bool          `mapstructure:"debugger_detection"`
	VMDetection       bool          `mapstructure:"vm_detection"`
	SleepOnDetection  time.Duration `mapstructure:"sleep_on_detection"`
}

// KillSwitchConfig contains kill switch settings
type KillSwitchConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	MaxNoContact  time.Duration `mapstructure:"max_no_contact"`
	AutoCleanup   bool          `mapstructure:"auto_cleanup"`
}