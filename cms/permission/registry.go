package permission

import (
	"slices"
	"strings"
	"sync"
)

type Registry struct {
	sync.RWMutex

	// map of internally used permission types to permission definitions
	definitions map[Type]Def

	// map mit den sprechenden Namen der Permissions/Definions
	// wird während der Generierung der Permission aus dem string in den Settings benutzt
	permissions map[string]Def
}

func (r *Registry) RegisterPermission(def Def) Type {

	r.RLock()
	defer r.RUnlock()

	def.Typ = Type(len(r.definitions)+1) << 2
	if def.Locale {
		def.Typ = def.Typ | typeWithLocaleMask
	}
	if def.Prefix {
		def.Typ = def.Typ | typeWithPrefixMask
	}

	r.definitions[def.Typ] = def
	name := strings.Join([]string{def.Namespace, def.Permission}, ".")
	r.permissions[name] = def

	return def.Typ
}

// GeneratePermission zerlegt permstr in permission, namespace, locales und prefix und generiert daraus eine Permission.
// Sollten Locales und prefix verlangt, aber nicht gefunden werden, werden diese aus den Argumenten 2 und 3 gelesen.
func (r *Registry) GeneratePermission(permstr string, locales []string, prefixes []string) (Type, Permission) {
	r.RLock()
	defer r.RUnlock()

	var perm Permission

	splits := strings.Split(permstr, ":")

	def, ok := r.permissions[splits[0]]
	if !ok {
		return 0, Permission{}
	}
	perm.ID = def.Permission
	perm.Namespace = def.Namespace
	if def.Typ.HasLocales() {

		if len(splits) > 1 {
			perm.Locales = strings.Split(splits[1], "|")
		} else {
			perm.Locales = locales
		}
	}
	if def.Typ.HasPrefixes() {
		if len(splits) > 2 {
			perm.Prefixes = strings.Split(splits[1], "|")
		} else {
			perm.Prefixes = prefixes
		}
	}
	return def.Typ, perm
}

type Type uint32

const typeWithLocaleMask Type = 1
const typeWithPrefixMask Type = 2

func (t Type) HasLocales() bool  { return t&typeWithLocaleMask != 0 }
func (t Type) HasPrefixes() bool { return t&typeWithPrefixMask != 0 }
func (t Type) String() string {
	d := globalregistry.definitions[t]
	// bytes, _ := d.MarshalText()
	return d.Namespace + "." + d.Permission
}

////////// package level registry //////////

var globalregistry Registry = Registry{definitions: map[Type]Def{}, permissions: map[string]Def{}}

func Define(namespace, name string, WithPrefix, WithLocale bool) Type {

	def := Def{Permission: name, Namespace: namespace, Prefix: WithPrefix, Locale: WithLocale}
	permtype := globalregistry.RegisterPermission(def)
	return permtype
}

func PermissionList() []string {
	var permissions []string
	for p := range globalregistry.permissions {
		permissions = append(permissions, p)
	}
	slices.Sort(permissions)
	return permissions
}
