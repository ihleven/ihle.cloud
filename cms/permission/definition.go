package permission

import "bytes"

type Def struct {
	Typ Type

	// Name is the name of the Permission
	Permission string
	Namespace  string `json:"namespace,omitempty"`

	// Ebenen:
	// * Art (Read,Write,...) -> aktuell nicht unterstützt, evtl. ein TODO
	// * Was (Folder, entries) -> Metadaten
	// * Locale (de-DE, *-*, *-CH, de-*, de)
	// * Pfadprefix z.B. alles unterhalb von /pages

	// Subset is a list of all allowed sub permissions, e.g. read, write
	// Subset []string

	// DefaultSubset is a list of sub permissions allowed when only the name of the permission is specified
	// DefaultSubset []string

	// permision settings
	Prefix bool `json:"prefix,omitempty"`
	Locale bool `json:"locales,omitempty"`
}

func (d Def) MarshalText() (text []byte, err error) {
	buffer := bytes.NewBufferString(d.Namespace + "." + d.Permission)
	if d.Locale {
		buffer.WriteString(":localized")
	}
	if d.Prefix {
		buffer.WriteString(":prefixed")
	}
	return buffer.Bytes(), nil
}
