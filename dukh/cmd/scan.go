package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
)

func scanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Trigger an immediate workspace rescan in dukh",
		RunE:  runScan,
	}
}

func runScan(_ *cobra.Command, _ []string) error {
	conn, client, err := dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ "+err.Error())
		return err
	}
	defer conn.Close()

	resp, err := client.Scan(context.Background(), &dukhpb.ScanRequest{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗ dukh scan failed: "+err.Error())
		return err
	}

	fmt.Println("✓ " + resp.Message)
	return nil
}
