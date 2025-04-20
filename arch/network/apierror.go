package network

import (
	"errors"
	"fmt"
	"net/http"
)

type apiError struct {
	Code    int
	Message string
	Err     error
}

func (e *apiError) GetCode() int {
	return e.Code
}

func (e *apiError) GetMessage() string {
	return e.Message
}

func (e *apiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%d - %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%d - %s", e.Code, e.Message)
}

func (e *apiError) Unwrap() error {
	return e.Err
}

func newApiError(code int, message string, err error) ApiError {
	apiError := apiError{
		Code:    code,
		Message: message,
		Err:     err,
	}
	if err == nil {
		apiError.Err = errors.New(message)
	}
	return &apiError
}

// 400 Bad Request - Malformed request, invalid input
func NewBadRequestError(message string, err error) ApiError {
	return newApiError(http.StatusBadRequest, message, err)
}

// 401 Unauthorized - Missing or invalid authentication
func NewUnauthorizedError(message string, err error) ApiError {
	return newApiError(http.StatusUnauthorized, message, err)
}

// 403 Forbidden - Valid auth but insufficient permissions
func NewForbiddenError(message string, err error) ApiError {
	return newApiError(http.StatusForbidden, message, err)
}

// 404 Not Found - Resource doesn't exist
func NewNotFoundError(message string, err error) ApiError {
	return newApiError(http.StatusNotFound, message, err)
}

// 405 Method Not Allowed - Wrong HTTP method
func NewMethodNotAllowedError(message string, err error) ApiError {
	return newApiError(http.StatusMethodNotAllowed, message, err)
}

// 406 Not Acceptable - Server can't fulfill requested format
func NewNotAcceptableError(message string, err error) ApiError {
	return newApiError(http.StatusNotAcceptable, message, err)
}

// 408 Request Timeout - Client took too long to send request
func NewRequestTimeoutError(message string, err error) ApiError {
	return newApiError(http.StatusRequestTimeout, message, err)
}

// 409 Conflict - Request conflicts with current state
func NewConflictError(message string, err error) ApiError {
	return newApiError(http.StatusConflict, message, err)
}

// 410 Gone - Resource no longer available
func NewGoneError(message string, err error) ApiError {
	return newApiError(http.StatusGone, message, err)
}

// 413 Payload Too Large - Request entity too large
func NewPayloadTooLargeError(message string, err error) ApiError {
	return newApiError(http.StatusRequestEntityTooLarge, message, err)
}

// 414 URI Too Long - Request URI too long
func NewURITooLongError(message string, err error) ApiError {
	return newApiError(http.StatusRequestURITooLong, message, err)
}

// 415 Unsupported Media Type - Incorrect Content-Type
func NewUnsupportedMediaTypeError(message string, err error) ApiError {
	return newApiError(http.StatusUnsupportedMediaType, message, err)
}

// 419 Authentication Timeout - Custom status for expired sessions
func NewSessionExpiredError(message string, err error) ApiError {
	return newApiError(419, message, err) // Custom status for session expiry
}

// 422 Unprocessable Entity - Semantic errors in request
func NewUnprocessableEntityError(message string, err error) ApiError {
	return newApiError(http.StatusUnprocessableEntity, message, err)
}

// 429 Too Many Requests - Rate limiting
func NewTooManyRequestsError(message string, err error) ApiError {
	return newApiError(http.StatusTooManyRequests, message, err)
}

// 500 Internal Server Error - Unexpected server error
func NewInternalServerError(message string, err error) ApiError {
	return newApiError(http.StatusInternalServerError, message, err)
}

// 501 Not Implemented - Feature not supported by server
func NewNotImplementedError(message string, err error) ApiError {
	return newApiError(http.StatusNotImplemented, message, err)
}

// 502 Bad Gateway - Invalid response from upstream server
func NewBadGatewayError(message string, err error) ApiError {
	return newApiError(http.StatusBadGateway, message, err)
}

// 503 Service Unavailable - Server temporarily unavailable
func NewServiceUnavailableError(message string, err error) ApiError {
	return newApiError(http.StatusServiceUnavailable, message, err)
}

// 504 Gateway Timeout - Upstream server timeout
func NewGatewayTimeoutError(message string, err error) ApiError {
	return newApiError(http.StatusGatewayTimeout, message, err)
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func NewHTTPVersionNotSupportedError(message string, err error) ApiError {
	return newApiError(http.StatusHTTPVersionNotSupported, message, err)
}
