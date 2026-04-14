package pkgman

// PkgLabel returns a display label for a package: "name" or "name@version".
func PkgLabel(pkg Package) string {
	if pkg.Version != "" {
		return pkg.Name + "@" + pkg.Version
	}
	return pkg.Name
}
