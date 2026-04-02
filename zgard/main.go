package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Zgard - The Command CLI")
		fmt.Println("Usage: zgard <command>")
		fmt.Println("Commands:")
		fmt.Println("  ws init    - Initialize workspace")
		fmt.Println("  ws purge   - Purge workspace")
		fmt.Println("  run        - Run a custom script")
		return
	}

	command := os.Args[1]
	switch command {
	case "ws":
		if len(os.Args) < 3 {
			fmt.Println("Usage: zgard ws <subcommand>")
			return
		}
		subcommand := os.Args[2]
		switch subcommand {
		case "init":
			fmt.Println("Initializing workspace...")
			// TODO: Implement workspace init
		case "purge":
			fmt.Println("Purging workspace...")
			// TODO: Implement workspace purge
		default:
			fmt.Printf("Unknown ws subcommand: %s\n", subcommand)
		}
	case "run":
		fmt.Println("Running custom script...")
		// TODO: Implement script running
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
