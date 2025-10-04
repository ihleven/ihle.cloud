package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

func (e *Error) Format(f fmt.State, verb rune) {

	switch verb {
	case 's':
		var str string
		if f.Flag('#') || f.Flag('+') {
			for _, a := range e.annotations {
				str += a.Message + ": "
			}
		}
		str += e.cause.Error()
		f.Write([]byte(str))

	case 'v':
		var bytes []byte
		var err error
		if f.Flag('#') || f.Flag('+') {
			bytes, err = json.MarshalIndent(e, "", "    ")
		} else {
			bytes, err = json.Marshal(e)
		}
		if err != nil {
			bytes = []byte(err.Error())
		}
		f.Write(bytes)
	}
}

func shortFuncName(f *runtime.Func) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	name := f.Name()
	return formatFuncname(name)
}

// "FuncName" or "Receiver.MethodName"
func formatFuncname(name string) string {
	// funcname removes the path prefix component of a function's name reported by func.Name().
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	name = name[i+1:]

	name = strings.Replace(name, "(", "", 1)
	name = strings.Replace(name, "*", "", 1)
	name = strings.Replace(name, ")", "", 1)
	return name
	// parts := strings.Split(name, "/")
	// return parts[len(parts)-1]
}

var Dir string

func formatFile(file string) string {

	file = removeLongestCommonPrefix(file, Dir)
	splits := strings.Split(file, "go/pkg/mod/")
	if len(splits) > 1 {
		file = "go/pkg/mod/" + splits[1]
	}

	return file
}

func removeLongestCommonPrefix(target, path string) string {
	for i := range target {
		if i >= len(path) {
			return target[i:]
		}

		if target[i] == path[i] {
			continue
		}

		return target[i:]
	}

	return target
}
