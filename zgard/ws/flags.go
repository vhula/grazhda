package ws

import (
"fmt"
"io"
"os"

clr "github.com/vhula/grazhda/internal/color"
"github.com/vhula/grazhda/internal/config"
"github.com/vhula/grazhda/internal/workspace"
)

// loadConfig resolves the config path, loads, and validates config.yaml.
// Any validation errors are written to stderr before the error is returned.
func loadConfig() (*config.Config, error) {
cfgPath := resolveConfigPath()
cfg, err := config.Load(cfgPath)
if err != nil {
return nil, err
}
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
