package content

import "bitbucket.org/hotelplan/webcc-content/cms/permission"

var (
	ENTRY_CREATE          = permission.Define("entry", "create", false, false)
	ENTRY_DELETE          = permission.Define("entry", "delete", false, false)
	ENTRY_META_SET_ID     = permission.Define("meta", "id.wrt", false, false)
	ENTRY_META_SET_PATH   = permission.Define("meta", "path.wrt", false, false)
	ENTRY_META_SET_LOCALE = permission.Define("meta", "locale.wrt", false, false)

	ENTRY_META_SET_ALL     = permission.Define("meta", "all", false, false)
	ENTRY_META_SET_NAME    = permission.Define("meta", "name.wrt", false, false)
	ENTRY_META_SET_NOTES   = permission.Define("meta", "notes.wrt", false, false)
	ENTRY_META_SET_TAGS    = permission.Define("meta", "tags.wrt", false, false)
	ENTRY_META_SET_STATUS  = permission.Define("meta", "status.wrt", false, false)
	ENTRY_META_SET_SLUG    = permission.Define("meta", "slug.wrt", false, false)
	ENTRY_META_SET_VERSION = permission.Define("meta", "version.wrt", false, false)

	ENTRY_META_SET_OWNER  = permission.Define("access", "owner.wrt", false, false)
	ENTRY_META_SET_GROUP  = permission.Define("access", "group.wrt", false, false)
	ENTRY_META_SET_PERMIS = permission.Define("access", "perm.wrt", false, false)
)
