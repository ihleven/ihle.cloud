package permission

import (
	"fmt"
	"slices"
	"strings"
)

type Scope map[Type]Permission

func (s Scope) Add(permissions []string, locales, paths []string) {
	for _, p := range permissions {

		t, permission := globalregistry.GeneratePermission(p, locales, paths)

		if t != 0 {
			s[t] = permission
		}
	}
}

// HasPermission checks if scope has permissions for a permission type and,
// if perm type has locales for a given locale to check for (in this case must be prefixed by one of the scopes locales), and,
// if perm type has paths for a given path to check for (must be prefixed by one of scopes paths).
func (s Scope) HasPermission(p Type, localeToCheck, pathToCheck string) bool {
	fmt.Printf(" +++++ HasPermission(%q, %q, %q) +++++\n", p, localeToCheck, pathToCheck)
	permission, ok := s[p]
	if !ok {
		return false
	}
	// fmt.Println(" +++++ PERMISSION +++++", p, localeToCheck, permission.Locales, pathToCheck, permission.Prefixes)

	if p.HasLocales() && len(permission.Locales) > 0 {
		allowed := false
		for _, l := range permission.Locales {
			fmt.Println("   +++ PERMISSION Locale+++++", localeToCheck, l)
			if strings.HasPrefix(localeToCheck, l) || l == "*" || l == "" {
				allowed = true
				break
			}
		}
		if !allowed {
			fmt.Printf(" +++++ Lacking Locale Permissions: %q -> %v +++++\n", localeToCheck, permission.Locales)
			return false
		}
	}
	if p.HasPrefixes() && len(permission.Prefixes) > 0 {

		allowed := false
		for _, permissionPath := range permission.Prefixes {
			fmt.Printf("   +++ PERMISSION Path+++++ %q %q\n", permissionPath, pathToCheck)
			if strings.HasPrefix(pathToCheck, permissionPath) || permissionPath == "*" || permissionPath == "" {
				allowed = true
				break
			}
		}
		if !allowed {
			fmt.Printf(" +++++ Lacking Path Permissions: %q -> %v +++++\n", pathToCheck, permission.Prefixes)

			return false
		}
	}
	return true
}

func (s Scope) HasPermissions(types ...Type) bool {
	for _, t := range types {
		if _, ok := s[t]; !ok {
			return false
		}
	}
	return true
}
func (s Scope) HasOnePermission(types ...Type) bool {
	for _, t := range types {
		if _, ok := s[t]; ok {
			return true
		}
	}
	return false
}

func (s Scope) Denies(permtype Type) bool {

	if _, ok := s[permtype]; !ok {
		return true
	}
	return false
}

func (s Scope) String() string {

	var data []string
	for _, perm := range s {
		data = append(data, perm.String())
	}
	slices.Sort(data)

	return strings.Join(data, ",")
}
