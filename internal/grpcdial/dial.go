package grpcdial

import (
	"fmt"

	"github.com/vhula/grazhda/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultHost = "localhost"
	defaultPort = 50501
)

// Addr returns the dukh server address from cfg, falling back to
// localhost:50501 when the config values are empty/zero.
func Addr(cfg *config.Config) string {
	host := cfg.Dukh.Host
	port := cfg.Dukh.Port
	if host == "" {
		host = defaultHost
	}
	if port == 0 {
		port = defaultPort
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// DefaultAddr returns the fallback address when no config is available.
func DefaultAddr() string {
	return fmt.Sprintf("%s:%d", defaultHost, defaultPort)
}

// Dial opens a lazy gRPC client connection to the given address.
// The caller is responsible for closing the returned connection.
func Dial(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create dukh client for %s: %w", addr, err)
	}
	return conn, nil
}
