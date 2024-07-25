package errors

type ApiError struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// var (
// 	INVALID_INPUT       = NewError(40000, "Invalid input")
// 	UNAUTHORIZED        = NewError(40001, "Unauthorized")
// 	FORBIDDEN           = NewError(40003, "Forbidden")
// 	NOT_FOUND           = NewError(40004, "Not Found")
// 	NOT_ACCEPTABLE      = NewError(40006, "Not Acceptable")
// 	SERVICE_UNAVAILABLE = NewError(99999, "Service unavailable, please try again later")
// )

const (
	INVALID_INPUT_CODE       = 40000
	UNAUTHORIZED_CODE        = 40001
	FORBIDDEN_CODE           = 40003
	NOT_FOUND_CODE           = 40004
	NOT_ACCEPTABLE_CODE      = 40006
	SERVICE_UNAVAILABLE_CODE = 99999

	INVALID_INPUT_CODE_MSG       = "Invalid input"
	UNAUTHORIZED_CODE_MSG        = "Unauthorized"
	FORBIDDEN_CODE_MSG           = "Forbidden"
	NOT_FOUND_CODE_MSG           = "Not Found"
	NOT_ACCEPTABLE_CODE_MSG      = "Not Acceptable"
	SERVICE_UNAVAILABLE_CODE_MSG = "Service unavailable, please try again later"
)

func (e *ApiError) Error() string {
	return e.Msg
}

func NewError(code int) *ApiError {
	var msg string
	switch code {
	case INVALID_INPUT_CODE:
		msg = INVALID_INPUT_CODE_MSG
	case UNAUTHORIZED_CODE:
		msg = UNAUTHORIZED_CODE_MSG
	case FORBIDDEN_CODE:
		msg = FORBIDDEN_CODE_MSG
	case NOT_FOUND_CODE:
		msg = NOT_FOUND_CODE_MSG
	case NOT_ACCEPTABLE_CODE:
		msg = NOT_ACCEPTABLE_CODE_MSG
	case SERVICE_UNAVAILABLE_CODE:
		msg = SERVICE_UNAVAILABLE_CODE_MSG
	}
	e := &ApiError{
		Msg:  msg,
		Code: code,
	}
	return e
}
