package pkgman

import (
	"fmt"
	"strings"
)

// Resolve returns a list of packages in topological (dependency-first) order.
// names selects which packages to include; if empty all registry packages are
// selected. The resolver always expands transitive dependencies automatically.
//
// Kahn's algorithm is used; a non-empty remainder after processing indicates a
// dependency cycle, which is reported with the names of the involved packages.
func Resolve(reg *Registry, names []string) ([]Package, error) {
	// Build a name→Package lookup.
	byName := make(map[string]Package, len(reg.Packages))
	for _, p := range reg.Packages {
		byName[p.Name] = p
	}

	// Collect the seed set; expand transitively to get the full closure.
	selected, err := closure(byName, names)
	if err != nil {
		return nil, err
	}

	// Build in-degree map and adjacency list over the selected set.
	inDegree := make(map[string]int, len(selected))
	deps := make(map[string][]string, len(selected)) // dependency → dependents
	for name := range selected {
		if _, seen := inDegree[name]; !seen {
			inDegree[name] = 0
		}
		for _, dep := range selected[name].DependsOn {
			depName, _ := parseDep(dep)
			inDegree[name]++
			deps[depName] = append(deps[depName], name)
		}
	}

	// Kahn's algorithm: seed queue with zero-in-degree nodes.
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	// Sort for deterministic output.
	sortStrings(queue)

	var ordered []Package
	for len(queue) > 0 {
		// Pop front.
		cur := queue[0]
		queue = queue[1:]
		ordered = append(ordered, selected[cur])

		for _, dependent := range deps[cur] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sortStrings(queue)
			}
		}
	}

	if len(ordered) != len(selected) {
		// Remaining non-zero in-degree nodes form the cycle.
		var cycle []string
		for name, deg := range inDegree {
			if deg > 0 {
				cycle = append(cycle, name)
			}
		}
		sortStrings(cycle)
		return nil, fmt.Errorf("dependency cycle detected among packages: %s", strings.Join(cycle, ", "))
	}

	return ordered, nil
}

// ResolveReverse returns packages in reverse topological order (dependents before
// their dependencies) for safe purge sequencing.
func ResolveReverse(reg *Registry, names []string) ([]Package, error) {
	ordered, err := Resolve(reg, names)
	if err != nil {
		return nil, err
	}
	// Reverse in place.
	for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
		ordered[i], ordered[j] = ordered[j], ordered[i]
	}
	return ordered, nil
}

// closure computes the transitive closure of the requested package names,
// verifying that every referenced package exists in the registry and that any
// version constraint specified in a depends_on entry is satisfied.
//
// depends_on entries may be plain package names ("sdkman") or versioned
// references ("sdkman@1.2.3"). A versioned reference is satisfied only when
// the registry package declares exactly that version.
func closure(byName map[string]Package, names []string) (map[string]Package, error) {
	// If no names given, select all.
	seeds := names
	if len(seeds) == 0 {
		for name := range byName {
			seeds = append(seeds, name)
		}
	}

	result := make(map[string]Package)
	stack := make([]string, len(seeds))
	copy(stack, seeds)

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// cur may itself be a versioned ref (when a dep pushes a versioned name
		// onto the stack); strip the version for the map key.
		curName, _ := parseDep(cur)

		if _, seen := result[curName]; seen {
			continue
		}
		p, ok := byName[curName]
		if !ok {
			return nil, fmt.Errorf("unknown package %q referenced in dependency chain", curName)
		}
		result[curName] = p
		for _, dep := range p.DependsOn {
			depName, depVer := parseDep(dep)
			if _, seen := result[depName]; !seen {
				stack = append(stack, depName)
			}
			// Validate version constraint even when the dep is already in the result.
			if depVer != "" {
				depPkg, known := byName[depName]
				if !known {
					return nil, fmt.Errorf("unknown package %q referenced by %q", depName, curName)
				}
				if depPkg.Version != depVer {
					return nil, fmt.Errorf(
						"package %q requires %s@%s but registry has %s@%s",
						curName, depName, depVer, depName, depPkg.Version,
					)
				}
			}
		}
	}
	return result, nil
}

// parseDep splits a depends_on entry into its package name and optional version.
// "sdkman"        → ("sdkman", "")
// "sdkman@1.2.3"  → ("sdkman", "1.2.3")
func parseDep(s string) (name, version string) {
	if idx := strings.Index(s, "@"); idx >= 0 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}

// sortStrings sorts a string slice in place using insertion sort
// (simple; small slices only).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
