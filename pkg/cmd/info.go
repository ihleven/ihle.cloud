package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

var Info info

type info struct {
	Go struct {
		Version string `json:"version"` // buildinfo.GoVersion
		Arch    string `json:"arch"`    // buildinfo.Settings GOARCH
		Os      string `json:"os"`      // buildinfo.Settings GOOS
	} `json:"go"`

	App struct {
		Name     string `json:"name"`
		Module   string `json:"module"`   // buildinfo.Path
		MainPath string `json:"mainpath"` // buildinfo.Main.Path
	} `json:"app"`

	Build struct {
		Dir     string `json:"dir"`     // befüllt aus LDFLAGS
		Time    string `json:"time"`    // befüllt aus LDFLAGS
		Output  string `json:"output"`  // befüllt aus LDFLAGS
		LDFlags string `json:"ldflags"` // buildinfo.Settings -ldflags
		Flags   string `json:"flags"`
		GitDesc string `json:"gitdesc"`
	} `json:"build"`
	Revision struct {
		Vcs       string    `json:"vcs"`       // befüllt aus debug.ReadBuildInfo() the version control system for the source tree where the build ran
		ID        string    `json:"id"`        // befüllt aus debug.ReadBuildInfo() the revision identifier for the current commit or checkout
		Timestamp time.Time `json:"timestamp"` // befüllt aus debug.ReadBuildInfo() the modification time associated with vcs.revision, in RFC3339 format
		Modified  bool      `json:"modified"`  // befüllt aus debug.ReadBuildInfo() true or false indicating whether the source tree had local modifications
	} `json:"revision"`

	Version version `json:"version"`
}

func (i *info) Dump() {
	bytes, _ := json.MarshalIndent(i, "", "    ")
	fmt.Println(string(bytes))
}

func New() info {
	buildinfo, _ := debug.ReadBuildInfo()

	info := info{}
	info.Go.Version = buildinfo.GoVersion
	info.App.Module = buildinfo.Path
	info.App.MainPath = buildinfo.Main.Path

	for _, s := range buildinfo.Settings {

		switch s.Key {
		case "GOARCH":
			info.Go.Arch = s.Value
		case "GOOS":
			info.Go.Os = s.Value
		case "-ldflags":
			info.Build.LDFlags = s.Value
		case "CGO_LDFLAGS":
			info.Build.Flags = s.Value
		case "vcs":
			info.Revision.Vcs = s.Value
		case "vcs.revision":
			info.Revision.ID = s.Value
		case "vcs.time":
			info.Revision.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", s.Value)
		case "vcs.modified":
			info.Revision.Modified, _ = strconv.ParseBool(s.Value)
		case "-buildmode", "-compiler", "CGO_ENABLED", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "GOARM64":
		default:
			fmt.Printf("unknown build info key: %s => %s\n", s.Key, s.Value)
		}
	}
	return info
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	response := struct {
		info
		Version string `json:"app_version"`
	}{
		Version: Info.Version.String(),
		info:    Info,
	}
	bytes, _ := json.MarshalIndent(response, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(bytes)
}

func SetLdflags(builddir, buildtime, buildoutput, desc string) {

	info := New()
	info.Build.Dir = builddir
	info.Build.Time = buildtime
	info.Build.Output = buildoutput
	info.Build.GitDesc = desc
	info.Version = VersionFromGitDescription(desc)
	// info.Dump()

	// fmt.Printf("VERSION: tag:%s major:%d minor:%d patch:%d revision:%s dirty:%t commits:%d prerelease:%s => %s\n", info.Version.tag, major, minor, patch, revision, dirty, commits, prerelease, info.Version)

	Info = info
}
