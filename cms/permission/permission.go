package permission

import (
	"strings"
)

// * Permissions sind strukturiert und haben eine Stringrepräsentierung, die äquivalent sind
// * Permissions werden in den Paketen definiert, in denen sie gebraucht werden (page_create in ctyp)
// * allg. Permissions sind in content oder permission direkt definioert
// * api/cms stellen mit jedem Request/Aktion die Permissions aus dem Context her (User, etc.)
// * owner / group werden über die Permissions IS_OWNER IS_IN_GROUP
//
// # Registrierung
// * Permissions sollen dort definiert werden wo sie verwendet werden, also in Paketen wie ctype oder content
// * Trotzdem müssen alle Permissions bekannt sein, wofür eine package level registry im permission Paket verwendet wird.
// * Beim Einlesen der Permission der Rollen und User werden nur permissions akzeptiert, die registiert sind.
// * CMS muss alle user kennen und wissen welche Permissions sie haben
//
// # package content
//
// * definiert generische Rollen
// * definiert Permissions auf Metadaten
// * defineirt generische Permissions
//
// # abgeleitete Packages
//
// * definiert welche Arten von permissions es gibt: ctype mit permission PAGE_CREATE
// * definiert evtl. zusätzliche Rollen samt zugehörigen Permissions
//
// # CMS Service
// * beim Startup werden Permissions und Rollen ausgelesen
//
// # Owner
//
// * Metadaten enthält Owner
// * Owner/Admin kann flag setzten: private oder: PermissionMode user/group/others -> R/W

type Permission struct {
	ID        string   `json:"id"`
	Namespace string   `json:"namespace,omitempty"`
	Prefixes  []string `json:"prefix,omitempty"`
	Locales   []string `json:"locales,omitempty"`
}

func (p *Permission) String() string {
	str := p.Namespace + "." + p.ID
	if len(p.Locales) > 0 {
		str += ":" + strings.Join(p.Locales, "|")
	}
	if len(p.Prefixes) > 0 {
		str += ":" + p.Prefixes[0]
		for _, path := range p.Prefixes[1:] {
			str += "|" + path
		}
	}
	return str
}
