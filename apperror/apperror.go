package apperror

import "fmt"

type Kind string

const (
	KindBadRequest         Kind = "bad_request"
	KindNotFound           Kind = "not_found"
	KindConflict           Kind = "conflict"
	KindUnprocessable      Kind = "unprocessable"
	KindRateLimited        Kind = "rate_limited"
	KindServiceUnavailable Kind = "service_unavailable"
)

type Error struct {
	Kind    Kind
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

func New(kind Kind, code, message string, cause error) *Error {
	return &Error{
		Kind:    kind,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func BadRequest(code, message string, cause error) *Error {
	return New(KindBadRequest, code, message, cause)
}

func NotFound(code, message string, cause error) *Error {
	return New(KindNotFound, code, message, cause)
}

func Conflict(code, message string, cause error) *Error {
	return New(KindConflict, code, message, cause)
}

func Unprocessable(code, message string, cause error) *Error {
	return New(KindUnprocessable, code, message, cause)
}

func RateLimited(code, message string, cause error) *Error {
	return New(KindRateLimited, code, message, cause)
}

func ServiceUnavailable(code, message string, cause error) *Error {
	return New(KindServiceUnavailable, code, message, cause)
}
