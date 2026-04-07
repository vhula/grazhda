package main

import (
	"context"
	"fmt"
	"time"

	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultAddr = "localhost:50501"

// dial opens a gRPC client connection to the dukh server.
// The caller is responsible for closing the returned connection.
func dial() (*grpc.ClientConn, dukhpb.DukhServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//nolint:staticcheck // DialContext is the correct API for this grpc version
	conn, err := grpc.DialContext(ctx, defaultAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot reach dukh at %s — is it running? (%w)", defaultAddr, err)
	}
	return conn, dukhpb.NewDukhServiceClient(conn), nil
}
