package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Backup copies cfgPath to cfgPath+".bak" byte-for-byte and returns the backup
// path. Backup is always called first; the .bak file persists even when the
// subsequent Replace or Merge fails, acting as a restore point.
func Backup(cfgPath string) (string, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", fmt.Errorf("reading config for backup: %w", err)
	}
	bak := cfgPath + ".bak"
	if err := os.WriteFile(bak, data, 0o600); err != nil {
		return "", fmt.Errorf("writing backup %s: %w", bak, err)
	}
	return bak, nil
}

// Replace parses and validates newData as a Config, then atomically writes it
// to cfgPath. Returns *ValidationError when the content is structurally invalid.
// Call Backup before Replace so a restore point exists.
func Replace(cfgPath string, newData []byte) error {
	var cfg Config
	if err := yaml.Unmarshal(newData, &cfg); err != nil {
		return fmt.Errorf("parsing replacement config: %w", err)
	}
	if errs := Validate(&cfg); len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}
	return atomicWrite(cfgPath, newData)
}

// Merge deep-merges patchData into the existing config at cfgPath. Maps are
// merged recursively; slices are replaced by the patch value when present.
// The merged result is validated before the atomic write.
// Call Backup before Merge so a restore point exists.
func Merge(cfgPath string, patchData []byte) error {
	baseData, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config for merge: %w", err)
	}

	var base map[string]interface{}
	if err := yaml.Unmarshal(baseData, &base); err != nil {
		return fmt.Errorf("parsing base config: %w", err)
	}
	if base == nil {
		base = map[string]interface{}{}
	}

	var patch map[string]interface{}
	if err := yaml.Unmarshal(patchData, &patch); err != nil {
		return fmt.Errorf("parsing patch config: %w", err)
	}

	merged := deepMergeMaps(base, patch)
	mergedData, err := yaml.Marshal(merged)
	if err != nil {
		return fmt.Errorf("serialising merged config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(mergedData, &cfg); err != nil {
		return fmt.Errorf("parsing merged config: %w", err)
	}
	if errs := Validate(&cfg); len(errs) > 0 {
		return &ValidationError{Errs: errs}
	}

	return atomicWrite(cfgPath, mergedData)
}

// deepMergeMaps merges patch into base. Map values are merged key-by-key;
// all other types (strings, numbers, booleans, slices) are replaced by the
// patch value.
func deepMergeMaps(base, patch map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, pv := range patch {
		bv, exists := result[k]
		if exists {
			if bMap, ok := bv.(map[string]interface{}); ok {
				if pMap, ok := pv.(map[string]interface{}); ok {
					result[k] = deepMergeMaps(bMap, pMap)
					continue
				}
			}
		}
		result[k] = pv
	}
	return result
}

// atomicWrite writes data to a temp file in the same directory as dst, then
// renames it into place. This ensures dst is never left partially written.
func atomicWrite(dst string, data []byte) error {
	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".zgard-cfg-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replacing %s: %w", dst, err)
	}
	return nil
}
