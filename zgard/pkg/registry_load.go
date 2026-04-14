package pkg

import (
	"fmt"

	"github.com/vhula/grazhda/internal/pkgman"
)

func loadMergedRegistry(grazhdaDir string) (*pkgman.Registry, error) {
	global, err := pkgman.LoadRegistry(pkgman.RegistryPath(grazhdaDir))
	if err != nil {
		return nil, fmt.Errorf("load global registry: %w", err)
	}
	local, err := pkgman.LoadLocalRegistry(pkgman.LocalRegistryPath(grazhdaDir))
	if err != nil {
		return nil, fmt.Errorf("load local registry: %w", err)
	}
	return pkgman.MergeRegistries(global, local), nil
}
