//go:build ignore

// gen-manpages generates Unix man pages for zgard into man/man1/.
// Run via: just man  (or: cd zgard && go run ../tools/gen-manpages/main.go)
package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"
	"github.com/vhula/grazhda/zgard/cfgcmd"
	"github.com/vhula/grazhda/zgard/ws"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "zgard",
		Short: "Workspace lifecycle manager",
	}
	root.AddCommand(ws.NewCmd())
	root.AddCommand(cfgcmd.NewCmd())

	outDir := "../man/man1"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", outDir, err)
	}

	header := &doc.GenManHeader{
		Title:   "ZGARD",
		Section: "1",
		Source:  "Grazhda",
		Manual:  "Grazhda Manual",
	}

	if err := doc.GenManTree(root, header, outDir); err != nil {
		log.Fatalf("gen man pages: %v", err)
	}
	log.Printf("man pages written to %s/", outDir)
}
