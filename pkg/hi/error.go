package hi

// Error is a failure with the HTTP status it should be reported as.
type Error struct {
	HTTPStatus int
	Code       string `json:"code"`
	Message    string `json:"msg"`
}

func (e Error) Error() string {
	return e.Message
}
