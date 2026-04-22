package client

import (
	"context"

	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"github.com/vhula/grazhda/internal/grpcdial"
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
