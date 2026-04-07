package main

import (
	"fmt"

	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultAddr = "localhost:50501"

// dial opens a gRPC client connection to the dukh server.
// The connection is lazy — it will fail on the first RPC if dukh is not running.
// The caller is responsible for closing the returned connection.
func dial() (*grpc.ClientConn, dukhpb.DukhServiceClient, error) {
	conn, err := grpc.NewClient(defaultAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create dukh client for %s: %w", defaultAddr, err)
	}
	return conn, dukhpb.NewDukhServiceClient(conn), nil
}
