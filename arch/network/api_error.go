package network

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

type apiError struct {
	StatusCode int
	ErrorCode  string
	Message    string
	Detail     string
	Err        error
}

const (
	// Error codes
	BadRequestErrorCode              = "BAD_REQUEST"
	UnauthorizedErrorCode            = "UNAUTHORIZED"
	ForbiddenErrorCode               = "FORBIDDEN"
	NotFoundErrorCode                = "NOT_FOUND"
	MethodNotAllowedErrorCode        = "METHOD_NOT_ALLOWED"
	NotAcceptableErrorCode           = "NOT_ACCEPTABLE"
	RequestTimeoutErrorCode          = "REQUEST_TIMEOUT"
	ConflictErrorCode                = "CONFLICT"
	GoneErrorCode                    = "GONE"
	PayloadTooLargeErrorCode         = "PAYLOAD_TOO_LARGE"
	URITooLongErrorCode              = "URI_TOO_LONG"
	UnsupportedMediaTypeErrorCode    = "UNSUPPORTED_MEDIA_TYPE"
	SessionExpiredErrorCode          = "SESSION_EXPIRED"
	UnprocessableEntityErrorCode     = "UNPROCESSABLE_ENTITY"
	TooManyRequestsErrorCode         = "TOO_MANY_REQUESTS"
	InternalServerErrorCode          = "INTERNAL_SERVER_ERROR"
	NotImplementedErrorCode          = "NOT_IMPLEMENTED"
	BadGatewayErrorCode              = "BAD_GATEWAY"
	ServiceUnavailableErrorCode      = "SERVICE_UNAVAILABLE"
	GatewayTimeoutErrorCode          = "GATEWAY_TIMEOUT"
	HTTPVersionNotSupportedErrorCode = "HTTP_VERSION_NOT_SUPPORTED"
	ErrorBindingCode                 = "BINDING_ERROR"
	ErrorValidationCode              = "VALIDATION_ERROR"
	ErrorFieldValidationCode         = "FIELD_VALIDATION_ERROR"
	UnknownErrorCode                 = "UNKNOWN_ERROR"

	DB_ERROR        = "DB_ERROR"
	CACHE_ERROR     = "CACHE_ERROR"
	FORBIDDEN_ERROR = "FORBIDDEN_ERROR"
	MEDIA_ERROR     = "MEDIA_ERROR"
)

func (e *apiError) GetStatusCode() int {
	return e.StatusCode
}

func (e *apiError) GetErrorCode() string {
	return e.ErrorCode
}

func (e *apiError) GetMessage() string {
	return e.Message
}

func IsApiError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*apiError)
	return ok
}

func AsApiError(err error) ApiError {
	// panic
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*apiError); ok {
		return apiErr
	}
	if apiErr, ok := err.(ApiError); ok {
		return apiErr
	}
	return &apiError{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  UnknownErrorCode,
		Message:    err.Error(),
		Err:        err,
	}
}

func (e *apiError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

func (e *apiError) Unwrap() error {
	if e.Err == nil {
		return errors.New(e.Message)
	}
	return e.Err
}

func (e *apiError) GetErrors(isDebug bool) []ErrorDetail {
	if e.Err == nil {
		return nil
	}

	var errors []ErrorDetail
	if isDebug {
		// Capture stacktrace, file, line, function
		stackBuf := make([]byte, 2048)
		stackSize := runtime.Stack(stackBuf, false)
		stacktrace := string(stackBuf[:stackSize])

		// Get file, line, function from the call stack
		var file, function string
		var line int
		if pc, f, l, ok := runtime.Caller(2); ok {
			file = f
			line = l
			function = runtime.FuncForPC(pc).Name()
		}

		errors = append(errors, NewErrorDetailWithDebug(
			e.ErrorCode,
			"",
			e.Message,
			fmt.Sprintf("Details: %s", e.Detail),
			fmt.Sprintf("Error: %s", e.Err.Error()),
			stacktrace,
			file,
			function,
			e.Err.Error(),
			line,
		))
	} else {
		errors = append(errors, NewErrorDetail(
			e.ErrorCode,
			"",
			e.Message,
			fmt.Sprintf("Details: %s", e.Detail),
			e.Err,
		))
	}

	return errors
}

func newApiError(statusCode int, message string, detail string, errCode string, err error) ApiError {
	apiError := apiError{
		StatusCode: statusCode,
		ErrorCode:  errCode,
		Message:    message,
		Detail:     detail,
		Err:        err,
	}
	if err == nil {
		apiError.Err = errors.New(message)
	}
	return &apiError
}

// 400 Bad Request - Malformed request, invalid input
func NewBadRequestError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusBadRequest, message, detail, BadRequestErrorCode, err)
}

// 401 Unauthorized - Missing or invalid authentication
func NewUnauthorizedError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusUnauthorized, message, detail, UnauthorizedErrorCode, err)
}

// 403 Forbidden - Valid auth but insufficient permissions
func NewForbiddenError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusForbidden, message, detail, ForbiddenErrorCode, err)
}

// 404 Not Found - Resource doesn't exist
func NewNotFoundError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusNotFound, message, detail, NotFoundErrorCode, err)
}

// 405 Method Not Allowed - Wrong HTTP method
func NewMethodNotAllowedError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusMethodNotAllowed, message, detail, MethodNotAllowedErrorCode, err)
}

// 406 Not Acceptable - Server can't fulfill requested format
func NewNotAcceptableError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusNotAcceptable, message, detail, NotAcceptableErrorCode, err)
}

// 408 Request Timeout - Client took too long to send request
func NewRequestTimeoutError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusRequestTimeout, message, detail, RequestTimeoutErrorCode, err)
}

// 409 Conflict - Request conflicts with current state
func NewConflictError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusConflict, message, detail, ConflictErrorCode, err)
}

// 410 Gone - Resource no longer available
func NewGoneError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusGone, message, detail, GoneErrorCode, err)
}

// 413 Payload Too Large - Request entity too large
func NewPayloadTooLargeError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusRequestEntityTooLarge, message, detail, PayloadTooLargeErrorCode, err)
}

// 414 URI Too Long - Request URI too long
func NewURITooLongError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusRequestURITooLong, message, detail, URITooLongErrorCode, err)
}

// 415 Unsupported Media Type - Incorrect Content-Type
func NewUnsupportedMediaTypeError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusUnsupportedMediaType, message, detail, UnsupportedMediaTypeErrorCode, err)
}

// 419 Authentication Timeout - Custom status for expired sessions
func NewSessionExpiredError(message string, detail string, err error) ApiError {
	return newApiError(419, message, detail, SessionExpiredErrorCode, err)
}

// 422 Unprocessable Entity - Semantic errors in request
func NewUnprocessableEntityError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusUnprocessableEntity, message, detail, UnprocessableEntityErrorCode, err)
}

// 429 Too Many Requests - Rate limiting
func NewTooManyRequestsError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusTooManyRequests, message, detail, TooManyRequestsErrorCode, err)
}

// 500 Internal Server Error - Unexpected server error
func NewInternalServerError(message string, detail string, errCode string, err error) ApiError {
	return newApiError(http.StatusInternalServerError, message, detail, errCode, err)
}

// 501 Not Implemented - Feature not supported by server
func NewNotImplementedError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusNotImplemented, message, detail, NotImplementedErrorCode, err)
}

// 502 Bad Gateway - Invalid response from upstream server
func NewBadGatewayError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusBadGateway, message, detail, BadGatewayErrorCode, err)
}

// 503 Service Unavailable - Server temporarily unavailable
func NewServiceUnavailableError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusServiceUnavailable, message, detail, ServiceUnavailableErrorCode, err)
}

// 504 Gateway Timeout - Upstream server timeout
func NewGatewayTimeoutError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusGatewayTimeout, message, detail, GatewayTimeoutErrorCode, err)
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func NewHTTPVersionNotSupportedError(message string, detail string, err error) ApiError {
	return newApiError(http.StatusHTTPVersionNotSupported, message, detail, HTTPVersionNotSupportedErrorCode, err)
}
