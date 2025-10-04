package errors

import (
	stderrors "errors"
	"fmt"
)

// New creates a new Error with given message as cause
func New(msg string, vals ...interface{}) error {

	return newError(fmt.Errorf(msg, vals...), 4)
}

// New creates a new Error with given message as cause, annotated with code
func NewWithCode(code int, msg string, vals ...interface{}) error {

	err := newError(fmt.Errorf(msg, vals...), 4)
	err.annotate(code, "")
	return err
}

// Wrap adds Annotation to existing error.
func Wrap(err error, msg string, vals ...interface{}) error {

	if err == nil {
		return nil
	}

	return wrap(err, 0, msg, vals...)
}

// Wrap adds Annotation to existing error.
func WrapWithCode(err error, code int, msg string, vals ...interface{}) error {

	if err == nil {
		return nil
	}

	return wrap(err, code, msg, vals...)
}

// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func Code(err error) int {

	var code int

	type coder interface{ Code() int }

	if errWithCode, ok := err.(coder); ok {
		code = errWithCode.Code()
	}
	return code
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}
