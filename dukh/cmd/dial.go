package main

import (
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"github.com/vhula/grazhda/internal/grpcdial"
	"google.golang.org/grpc"
)

// dial opens a gRPC client connection to the dukh server using the default
// address. The caller is responsible for closing the returned connection.
func dial() (*grpc.ClientConn, dukhpb.DukhServiceClient, error) {
	addr := grpcdial.DefaultAddr()
	conn, err := grpcdial.Dial(addr)
	if err != nil {
		return nil, nil, err
	}
	return conn, dukhpb.NewDukhServiceClient(conn), nil
}
