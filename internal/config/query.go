package config

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetByPath retrieves a value from cfg at the given dotted-path key and
// returns it serialised as a string.
//
//   - Scalar values (string, int, bool, float) are returned as their natural
//     string representation.
//   - Complex values (maps, sequences) are returned as compact YAML.
//
// Key segments are joined with dots; array elements are addressed by
// zero-based integer index:
//
//	"editor"               → the editor string
//	"dukh.port"            → the dukh port as a string
//	"workspaces.0.name"    → the name of the first workspace
func GetByPath(cfg *Config, key string) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("serialising config: %w", err)
	}

	var root interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return "", fmt.Errorf("parsing config: %w", err)
	}

	val, err := traversePath(root, strings.Split(key, "."), key)
	if err != nil {
		return "", err
	}

	return formatValue(val)
}

// Serialize returns the full configuration serialised as YAML.
func Serialize(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// traversePath walks cur along the dotted-path segments.
func traversePath(cur interface{}, parts []string, fullKey string) (interface{}, error) {
	if len(parts) == 0 {
		return cur, nil
	}
	part := parts[0]
	rest := parts[1:]

	switch v := cur.(type) {
	case map[string]interface{}:
		val, ok := v[part]
		if !ok {
			return nil, fmt.Errorf("key %q not found in configuration", fullKey)
		}
		return traversePath(val, rest, fullKey)
	case []interface{}:
		idx, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("expected array index at segment %q in key %q", part, fullKey)
		}
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("index %d out of range (len=%d) at key %q", idx, len(v), fullKey)
		}
		return traversePath(v[idx], rest, fullKey)
	default:
		return nil, fmt.Errorf("cannot traverse into %T at segment %q (full key: %q)", cur, part, fullKey)
	}
}

// formatValue converts a generic value to a printable string.
func formatValue(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case int:
		return strconv.Itoa(val), nil
	case bool:
		return strconv.FormatBool(val), nil
	case float64:
		return fmt.Sprintf("%g", val), nil
	default:
		out, err := yaml.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val), nil
		}
		return strings.TrimRight(string(out), "\n"), nil
	}
}
