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

// StatusCode reports the status this error should be answered with.
//
// The status is a field rather than a method because it is data the store's
// reply carried; this is the method that makes it readable by code that must
// not know what a HiDrive error is — the HTTP layer asks every error it is
// handed whether it carries a status, and this is how this one answers.
//
// It is also the shape errs.StatusOf looks for, so an error from here gets the
// right status wherever that is used too.
func (e Error) StatusCode() int {
	return e.HTTPStatus
}
