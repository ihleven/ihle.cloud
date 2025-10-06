package hi

import (
	"fmt"
)

// type Error struct {
// 	Status  int    `json:"-"`   // HTTP status
// 	Message string `json:"msg"` // The msg message is intended to support developers during development and debugging. It is recommended to avoid automated parsing, since its structure is optimized for human readability.
// 	Code_   int    `json:"code"`
// }

// func (e *Error) Error() string {
// 	bytes, _ := json.MarshalIndent(e, "", "    ")
// 	return fmt.Sprintf("%s", bytes)
// }

// func (e *Error) Code() int {
// 	return e.Status
// }

func newerr(msg string, args ...interface{}) Error {
	return Error{Message: fmt.Sprintf(msg, args)}
}

type Error struct {
	HTTPStatus int
	Code       string `json:"code"`
	Message    string `json:"msg"`
}

func (e Error) Error() string {
	return e.Message
}
