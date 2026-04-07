package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
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
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}
	defer conn.Close()

	resp, err := client.Scan(context.Background(), &dukhpb.ScanRequest{})
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ dukh scan failed: "+err.Error()))
		return err
	}

	fmt.Println(icolor.Green("✓ " + resp.Message))
	return nil
}
