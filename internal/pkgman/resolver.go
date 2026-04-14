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
	idx, err := newPackageIndex(reg)
	if err != nil {
		return nil, err
	}

	// Collect the seed set; expand transitively to get the full closure.
	selected, err := closure(idx, names)
	if err != nil {
		return nil, err
	}

	// Build in-degree map and adjacency list over the selected set.
	inDegree := make(map[string]int, len(selected))  // package id -> dep count
	deps := make(map[string][]string, len(selected)) // dependency id -> dependent ids
	for id, pkg := range selected {
		if _, seen := inDegree[id]; !seen {
			inDegree[id] = 0
		}
		for _, dep := range pkg.DependsOn {
			depID, _, err := resolvePackageRef(idx, dep)
			if err != nil {
				return nil, fmt.Errorf("package %q dependency %q: %w", pkgID(pkg), dep, err)
			}
			inDegree[id]++
			deps[depID] = append(deps[depID], id)
		}
	}

	// Kahn's algorithm: seed queue with zero-in-degree nodes.
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
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
		for id, deg := range inDegree {
			if deg > 0 {
				cycle = append(cycle, id)
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

// closure computes the transitive closure of the requested package refs,
// verifying that every referenced package exists in the registry and that any
// version constraint specified in a depends_on entry is satisfied.
//
// depends_on entries may be plain package names ("sdkman") or versioned
// references ("sdkman@1.2.3"). A versioned reference is satisfied only when
// the registry package declares exactly that version.
func closure(idx *packageIndex, refs []string) (map[string]Package, error) {
	seeds := refs
	if len(seeds) == 0 {
		for id := range idx.byID {
			seeds = append(seeds, id)
		}
	}

	result := make(map[string]Package)
	stack := make([]string, len(seeds))
	copy(stack, seeds)

	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		curID, pkg, err := resolvePackageRef(idx, cur)
		if err != nil {
			return nil, fmt.Errorf("unknown package %q referenced in dependency chain", cur)
		}
		if _, seen := result[curID]; seen {
			continue
		}
		result[curID] = pkg
		for _, dep := range pkg.DependsOn {
			depID, _, err := resolvePackageRef(idx, dep)
			if err != nil {
				return nil, fmt.Errorf("package %q dependency %q: %w", pkgID(pkg), dep, err)
			}
			if _, seen := result[depID]; !seen {
				stack = append(stack, depID)
			}
		}
	}
	return result, nil
}

type packageIndex struct {
	byID   map[string]Package
	byName map[string][]Package
}

func newPackageIndex(reg *Registry) (*packageIndex, error) {
	idx := &packageIndex{
		byID:   make(map[string]Package, len(reg.Packages)),
		byName: make(map[string][]Package, len(reg.Packages)),
	}
	for _, p := range reg.Packages {
		id := pkgID(p)
		if _, exists := idx.byID[id]; exists {
			return nil, fmt.Errorf("duplicate package entry %q in registry", id)
		}
		idx.byID[id] = p
		idx.byName[p.Name] = append(idx.byName[p.Name], p)
	}
	return idx, nil
}

func resolvePackageRef(idx *packageIndex, ref string) (string, Package, error) {
	name, version := parseDep(ref)
	if version != "" {
		id := pkgID(Package{Name: name, Version: version})
		pkg, ok := idx.byID[id]
		if !ok {
			return "", Package{}, fmt.Errorf("required version %q for package %q is not defined", version, name)
		}
		return id, pkg, nil
	}

	candidates := idx.byName[name]
	if len(candidates) == 0 {
		return "", Package{}, fmt.Errorf("package %q is not defined", name)
	}
	if len(candidates) == 1 {
		pkg := candidates[0]
		return pkgID(pkg), pkg, nil
	}

	for _, pkg := range candidates {
		if pkg.Version == "" {
			return pkgID(pkg), pkg, nil
		}
	}
	return "", Package{}, fmt.Errorf("package %q has multiple versions; use %q in depends_on", name, name+"@<version>")
}

func pkgID(p Package) string {
	if p.Version == "" {
		return p.Name
	}
	return p.Name + "@" + p.Version
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
