package pkg

import "fmt"

// validateNameOrAll checks that exactly one of --name or --all is provided.
func validateNameOrAll(name string, all bool) error {
	if !all && name == "" {
		return fmt.Errorf("provide --name <pkg> or --all")
	}
	if all && name != "" {
		return fmt.Errorf("--all and --name are mutually exclusive")
	}
	return nil
}
