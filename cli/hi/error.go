package hi

import "fmt"

type Error struct {
	Status  int    `json:"-"`   // HTTP status
	Message string `json:"msg"` // The msg message is intended to support developers during development and debugging. It is recommended to avoid automated parsing, since its structure is optimized for human readability.

	Code string `json:"code"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("{%d: %s}", e.Code, e.Message)
}
