package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Dukh - The Worker CLI")
		fmt.Println("Usage: dukh <command>")
		fmt.Println("Commands:")
		fmt.Println("  start    - Start the Dukh server")
		fmt.Println("  stop     - Stop the Dukh server")
		fmt.Println("  status   - Check server status")
		return
	}

	command := os.Args[1]
	switch command {
	case "start":
		fmt.Println("Starting Dukh server...")
		// TODO: Implement server start logic
	case "stop":
		fmt.Println("Stopping Dukh server...")
		// TODO: Implement server stop logic
	case "status":
		fmt.Println("Dukh server status: Not implemented yet")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
