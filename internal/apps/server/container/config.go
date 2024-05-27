package container

import (
	"net"
	"time"
)

// Config holds the configuration for the server application.
// It includes settings for logging, storage, database connection, and other application parameters.
type Config struct {
	LogLevel        string        // LogLevel specifies the logging level (e.g., "debug", "info", "warn", "error").
	StoreInterval   time.Duration // StoreInterval defines the interval at which metrics are stored.
	FileStoragePath string        // FileStoragePath is the path to the directory where metrics are stored in file storage.
	Restore         bool          // Restore indicates whether to restore metrics from storage on startup.
	DatabaseDSN     string        // DatabaseDSN is the data source name for connecting to the database.
	Key             string        // Key is the secret key used for hashing.
	Addr            string        // Server host and port.
	GRPCAddr        string        // GRPCServer host and port.
	PrivateKeyPath  string        // Path to private PEM key for decrypting incoming metrics.
	TrustedSubnet   *net.IPNet    // Trusted subnet (CIDR)
}
