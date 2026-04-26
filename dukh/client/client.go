package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/grpcdial"
	ipath "github.com/vhula/grazhda/internal/path"
	"google.golang.org/grpc"
)

// Client is a typed gRPC client for the dukh health monitor.
type Client struct {
	conn *grpc.ClientConn
	svc  dukhpb.DukhServiceClient
}

// New dials addr and returns a ready Client. The caller must call Close when done.
func New(addr string) (*Client, error) {
	conn, err := grpcdial.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, svc: dukhpb.NewDukhServiceClient(conn)}, nil
}

// NewDefault dials the default localhost:50501 address.
func NewDefault() (*Client, error) {
	return New(grpcdial.DefaultAddr())
}

// Connect creates a Client using the address from cfg. If dukh is not reachable
// and cfg.Dukh.Autostart is true, Connect starts the daemon, waits for it to
// become ready, and then returns the client.
func Connect(cfg *config.Config) (*Client, error) {
	addr := grpcdial.Addr(cfg)
	c, err := New(addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	_, probeErr := c.svc.Status(ctx, &dukhpb.StatusRequest{})
	cancel()

	if probeErr != nil {
		if !cfg.Dukh.Autostart {
			c.conn.Close()
			return nil, fmt.Errorf("dukh is not running (autostart is disabled)")
		}
		if err := c.ensureRunning(); err != nil {
			c.conn.Close()
			return nil, err
		}
	}

	return c, nil
}

// ensureRunning starts the dukh daemon and waits for it to become ready.
func (c *Client) ensureRunning() error {
	fmt.Println(color.Blue("⟳ dukh is not running — starting…"))

	dukhBin := ipath.DukhBin()
	cmd := exec.Command(dukhBin, "start")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("auto-start dukh: %w", err)
	}

	if err := c.waitForReady(10 * time.Second); err != nil {
		return fmt.Errorf("dukh did not become ready: %w", err)
	}

	fmt.Println(color.Green("✓") + " dukh started")
	fmt.Println()
	return nil
}

// waitForReady polls the gRPC server until a Status RPC succeeds or the timeout elapses.
func (c *Client) waitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err := c.svc.Status(ctx, &dukhpb.StatusRequest{})
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s", timeout)
}

// Status queries workspace health. Pass an empty wsName for all workspaces.
func (c *Client) Status(ctx context.Context, wsName string, rescan bool) (*dukhpb.StatusResponse, error) {
	return c.svc.Status(ctx, &dukhpb.StatusRequest{WorkspaceName: wsName, Rescan: rescan})
}

// Stop sends a graceful shutdown request and returns the daemon's confirmation message.
func (c *Client) Stop(ctx context.Context) (string, error) {
	resp, err := c.svc.Stop(ctx, &dukhpb.StopRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

// Scan triggers an immediate workspace rescan and returns the daemon's confirmation message.
func (c *Client) Scan(ctx context.Context) (string, error) {
	resp, err := c.svc.Scan(ctx, &dukhpb.ScanRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
