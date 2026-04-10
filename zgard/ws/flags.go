package ws

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/workspace"
)

// loadConfig resolves the config path, loads, applies env overrides, and
// validates config.yaml.  Any validation errors are written to stderr before
// the error is returned.
func loadConfig() (*config.Config, error) {
	cfgPath := resolveConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(
				"configuration file not found at %s\n\n"+
					"  To create one:\n"+
					"    cp config.template.yaml %s",
				cfgPath, cfgPath,
			)
		}
		return nil, err
	}
	config.ApplyEnvOverrides(cfg)
	if errs := config.Validate(cfg); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, clr.Red(e))
		}
		return nil, fmt.Errorf("configuration is invalid")
	}
	return cfg, nil
}

// warnDefaultTarget prints an info message to out when the user has not
// provided an explicit targeting flag and the command falls back to the
// default workspace.
func warnDefaultTarget(out io.Writer, ws config.Workspace) {
	fmt.Fprintln(out, clr.Blue(fmt.Sprintf(
		"Info: Targeting default workspace: %s",
		workspace.ExpandHome(ws.Path),
	)))
}

// rootFlag reads a bool persistent flag from the root command.
// Returns false (zero value) if the flag is not defined.
func rootFlag(cmd *cobra.Command, name string) bool {
	root := cmd.Root()
	if f := root.PersistentFlags().Lookup(name); f != nil {
		return f.Value.String() == "true"
	}
	return false
}
