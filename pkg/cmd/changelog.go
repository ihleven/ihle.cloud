package cmd

// termmarkdown "github.com/MichaelMure/go-term-markdown"
// "github.com/gomarkdown/markdown"
// "github.com/gomarkdown/markdown/html"

var CHANGELOG string

// func Changelog() {

// 	result := termmarkdown.Render(string(CHANGELOG), 100, 0)

// 	fmt.Printf("%s\n", result)
// }

// func ChangelogHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Header.Get("Accept") == "text/markdown" {
// 		w.Write([]byte(CHANGELOG))
// 		return
// 	}
// 	ast := markdown.Parse([]byte(CHANGELOG), nil)
// 	fmt.Fprintf(w, "%s", markdown.Render(ast, html.NewRenderer(html.RendererOptions{})))
// }

// func SetChangelog(content string) {
// 	CHANGELOG = content
// 	// Version = parseVersionFromChangelog(content)
// }

// func parseVersionFromChangelog(CHANGELOG string) (v version) {

// 	re := regexp.MustCompile(`(?m)^##.*\[(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)[^a-z]*?(?P<milestone>[a-z]*)\s*$`)
// 	match := re.FindStringSubmatch(CHANGELOG)
// 	if len(match) > 0 {

// 		result := make(map[string]string)
// 		for i, name := range re.SubexpNames() {
// 			if i != 0 && name != "" {
// 				result[name] = match[i]
// 			}
// 		}

// 		if len(match) > 2 {
// 			v.Minor, _ = strconv.Atoi(match[2])
// 		}
// 		if len(match) > 3 {
// 			v.Patch, _ = strconv.Atoi(match[3])
// 		}

// 		if major, ok := result["major"]; ok {
// 			v.Major, _ = strconv.Atoi(major)
// 		}

// 		if minor, ok := result["minor"]; ok {
// 			v.Minor, _ = strconv.Atoi(minor)
// 		}

// 		if patch, ok := result["patch"]; ok {
// 			v.Patch, _ = strconv.Atoi(patch)
// 		}
// 		if milestone, ok := result["milestone"]; ok {
// 			v.PreReleaseMilestone = milestone
// 		}

// 	}
// 	return
// }
