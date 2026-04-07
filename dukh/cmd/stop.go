package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
)

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running dukh workspace monitor",
		RunE:  runStop,
	}
}

func runStop(_ *cobra.Command, _ []string) error {
	conn, client, err := dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ "+err.Error())
		return err
	}
	defer conn.Close()

	resp, err := client.Stop(context.Background(), &dukhpb.StopRequest{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ dukh stop failed: "+err.Error())
		return err
	}

	fmt.Println("✓ " + resp.Message)
	return nil
}
