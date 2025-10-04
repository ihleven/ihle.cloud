package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
)

func callers(skip int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// stack represents a stack of program counters.
type stack []uintptr

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the function for this Frame's pc.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the function for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of this function, if known.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

func (f Frame) fileLineFunc() (string, int, string) {

	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown file", 0, "unknown function"
	}

	file, line := fn.FileLine(f.pc())

	return formatFile(file), line, shortFuncName(fn)
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	fmt.Printf("Frame.Format(%c)\n", verb)
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			_, _ = io.WriteString(s, f.name())
			_, _ = io.WriteString(s, "\n\t")
			_, _ = io.WriteString(s, f.file())
		default:
			_, _ = io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		_, _ = io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		_, _ = io.WriteString(s, formatFuncname(f.name()))
	case 'v':
		f.Format(s, 's')
		_, _ = io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

/////////////////////////////////

// Frame is a single step in stack trace.
type StackFrameDep struct {
	Func string
	Line int
	Path string
}

func (f StackFrameDep) String() string {
	// return fmt.Sprintf("%s:%d %s()", f.Path, f.Line, f.Func)
	return fmt.Sprintf("%s (%s:%d)", f.Func, f.Path, f.Line)
}

//nolint:unused
func traceDep(skip int) []StackFrameDep {

	frames := make([]StackFrameDep, 0, 64)
	for {
		pc, path, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		frame := StackFrameDep{
			Func: shortFuncName(fn), //fn.Name(),
			Line: line,
			Path: formatFile(path),
		}
		frames = append(frames, frame)
		skip++
	}
	return frames
}
