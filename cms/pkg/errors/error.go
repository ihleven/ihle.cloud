package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
)

func wrap(cause error, code int, msg string, vals ...interface{}) *Error {

	e, ok := cause.(*Error)
	if !ok {
		e = newError(cause, 5)
	}

	e.annotate(code, msg, vals...)

	return e
}

func newError(cause error, skip int) *Error {

	return &Error{
		cause:      cause,
		stacktrace: callers(skip),
	}
}

type Error struct {
	cause       error
	annotations []annotation
	stacktrace  *stack // stacktrace    []StackFrame
}

func (e *Error) Error() string {
	return e.cause.Error()
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Code() int {
	for _, a := range e.annotations {
		if a.Code != 0 {
			return a.Code
		}
	}
	type coder interface{ Code() int }
	if ewc, ok := e.cause.(coder); ok {
		return ewc.Code()
	}
	return 500
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) annotate(code int, msg string, vals ...interface{}) {

	st := annotation{
		Code:    code,
		Message: fmt.Sprintf(msg, vals...),
	}

	pc, file, line, ok := runtime.Caller(3)
	if ok {
		st.File, st.Line = formatFile(file), line

		f := runtime.FuncForPC(pc)
		if f != nil {
			st.Function = shortFuncName(f)
		}
	}

	e.annotations = append([]annotation{st}, e.annotations...)
}

type annotation struct {
	Message  string
	Code     int
	File     string
	Line     int
	Function string
}

func (st *annotation) MarshalJSON() ([]byte, error) {

	tmp := struct {
		Code int    `json:"code,omitempty"`
		Msg  string `json:"msg,omitempty"`
		At   string `json:"at,omitempty"`
	}{
		Code: st.Code,
		Msg:  st.Message,
		At:   fmt.Sprintf("%s:%d (%s)", st.File, st.Line, st.Function),
	}

	return json.Marshal(tmp)
}

// MarshalJSON implements stdlib json interface
// This is used for formatting the error with the 'v' verb
func (e *Error) MarshalJSON() ([]byte, error) {

	foo := struct {
		Annotations []annotation `json:"annotations,omitempty"`
		Cause       string       `json:"cause,omitempty"`
		Stacktrace  []string     `json:"stacktrace,omitempty"`
	}{
		Annotations: e.annotations,
		Cause:       e.cause.Error(),
	}

	if e.stacktrace != nil {

		for _, s := range *e.stacktrace {

			file, line, function := Frame(s).fileLineFunc()
			foo.Stacktrace = append(foo.Stacktrace,
				file+":"+strconv.Itoa(line)+" ("+function+")",
			)
		}
	}

	return json.Marshal(foo)
}
