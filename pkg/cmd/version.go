package cmd

import (
	"fmt"
	"regexp"
	"strconv"
)

func Version() string {
	version := Info.Version.String()

	if Info.Version.IsDirty {
		version += " (uncommitted changes)"
	}

	return "tool-api " + version
}

type version struct {
	Major               int    `json:"major"`
	Minor               int    `json:"minor"`
	Patch               int    `json:"patch"`
	PreReleaseMilestone string `json:"prerelease_milestone"`
	PreReleaseNumber    int    `json:"prerelease_number"` // 5
	Branch              string `json:"-"`
	Description         string `json:"description"` // git describe --always --long --dirty => v2.9.1-5-g305e836-dirty
	Tag                 string `json:"-"`           // v2.9.1
	VCS                 string `json:"vcs"`         // git
	Revision            string `json:"revision"`    // g305e836
	IsDirty             bool   `json:"dirty"`       // true
}

func (v version) String() string {
	version := fmt.Sprintf("v%d", v.Major)
	if v.Minor != 0 {
		version += fmt.Sprintf(".%d", v.Minor)
	}
	if v.Patch != 0 {
		version += fmt.Sprintf(".%d", v.Patch)
	}
	if v.PreReleaseMilestone != "" {
		version += "-" + v.PreReleaseMilestone
	}
	if v.PreReleaseNumber != 0 {
		version += "." + strconv.Itoa(v.PreReleaseNumber)
	}
	if v.Revision != "" {
		version += "-" + v.Revision
	}
	if v.IsDirty {
		version += "-dirty"
	}
	return version
}

// parses tag, commits ahead, revision hash and dirty status from a string produced with
// git describe --always --long --dirty (e.g. v2.9.1-5-g305e836-dirty)
func VersionFromGitDescription(desc string) version {

	var (
		tag                 string
		major, minor, patch uint64
		commits             int
		revision            string
		isdirty             bool
		prerelease          string
	)
	// var version, revision, commits, dirty string
	var rgxVersion = regexp.MustCompile(`^(v([0-9]+)(\.([0-9]+))?(\.([0-9]+))?)(\-([_\w]+))?(\-([0-9]+)\-g([0-9a-f]+))?(\-(.*))?$`)

	matches := rgxVersion.FindStringSubmatch(desc)
	if len(matches) > 0 {
		tag = matches[1]
		major, _ = strconv.ParseUint(matches[2], 10, 64)
		minor, _ = strconv.ParseUint(matches[4], 10, 64)
		patch, _ = strconv.ParseUint(matches[6], 10, 64)
		prerelease = matches[8]
		if matches[10] != "" {
			commits, _ = strconv.Atoi(matches[10]) //fmt.Sprintf(" %s commits ahead of", matches[3])
			if matches[11] != "" {
				revision = matches[11]
				if matches[13] != "" {
					isdirty = true //" (has uncommitted changes)"
				}
			}
		}
	}

	return version{
		Major:               int(major),
		Minor:               int(minor),
		Patch:               int(patch),
		PreReleaseMilestone: prerelease,
		PreReleaseNumber:    commits,
		Description:         desc,
		Tag:                 tag,
		Revision:            revision,
		IsDirty:             isdirty,
	}
}
