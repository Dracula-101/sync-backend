package network

import (
	"errors"
	"fmt"
	"net/http"
)

type apiError struct {
	StatusCode int
	ErrorCode  string
	Message    string
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
	ErrorValidationCode              = "VALIDATION_ERROR"
	ErrorFieldValidationCode         = "FIELD_VALIDATION_ERROR"
	UnknownErrorCode                 = "UNKNOWN_ERROR"
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
		errors = append(errors, ErrorDetail{
			Code:    e.ErrorCode,
			Message: e.Message,
			Detail:  e.Err.Error(),
		})
	} else {
		errors = append(errors, ErrorDetail{
			Code:    e.ErrorCode,
			Message: e.Message,
			Detail:  "An error occurred. Please try again later.",
		})
	}

	return errors
}

func newApiError(statusCode int, message string, errCode string, err error) ApiError {
	apiError := apiError{
		StatusCode: statusCode,
		ErrorCode:  errCode,
		Message:    message,
		Err:        err,
	}
	if err == nil {
		apiError.Err = errors.New(message)
	}
	return &apiError
}

// 400 Bad Request - Malformed request, invalid input
func NewBadRequestError(message string, err error) ApiError {
	return newApiError(http.StatusBadRequest, message, BadRequestErrorCode, err)
}

// 401 Unauthorized - Missing or invalid authentication
func NewUnauthorizedError(message string, err error) ApiError {
	return newApiError(http.StatusUnauthorized, message, UnauthorizedErrorCode, err)
}

// 403 Forbidden - Valid auth but insufficient permissions
func NewForbiddenError(message string, err error) ApiError {
	return newApiError(http.StatusForbidden, message, ForbiddenErrorCode, err)
}

// 404 Not Found - Resource doesn't exist
func NewNotFoundError(message string, err error) ApiError {
	return newApiError(http.StatusNotFound, message, NotFoundErrorCode, err)
}

// 405 Method Not Allowed - Wrong HTTP method
func NewMethodNotAllowedError(message string, err error) ApiError {
	return newApiError(http.StatusMethodNotAllowed, message, MethodNotAllowedErrorCode, err)
}

// 406 Not Acceptable - Server can't fulfill requested format
func NewNotAcceptableError(message string, err error) ApiError {
	return newApiError(http.StatusNotAcceptable, message, NotAcceptableErrorCode, err)
}

// 408 Request Timeout - Client took too long to send request
func NewRequestTimeoutError(message string, err error) ApiError {
	return newApiError(http.StatusRequestTimeout, message, RequestTimeoutErrorCode, err)
}

// 409 Conflict - Request conflicts with current state
func NewConflictError(message string, err error) ApiError {
	return newApiError(http.StatusConflict, message, ConflictErrorCode, err)
}

// 410 Gone - Resource no longer available
func NewGoneError(message string, err error) ApiError {
	return newApiError(http.StatusGone, message, GoneErrorCode, err)
}

// 413 Payload Too Large - Request entity too large
func NewPayloadTooLargeError(message string, err error) ApiError {
	return newApiError(http.StatusRequestEntityTooLarge, message, PayloadTooLargeErrorCode, err)
}

// 414 URI Too Long - Request URI too long
func NewURITooLongError(message string, err error) ApiError {
	return newApiError(http.StatusRequestURITooLong, message, URITooLongErrorCode, err)
}

// 415 Unsupported Media Type - Incorrect Content-Type
func NewUnsupportedMediaTypeError(message string, err error) ApiError {
	return newApiError(http.StatusUnsupportedMediaType, message, UnsupportedMediaTypeErrorCode, err)
}

// 419 Authentication Timeout - Custom status for expired sessions
func NewSessionExpiredError(message string, err error) ApiError {
	return newApiError(419, message, SessionExpiredErrorCode, err)
}

// 422 Unprocessable Entity - Semantic errors in request
func NewUnprocessableEntityError(message string, err error) ApiError {
	return newApiError(http.StatusUnprocessableEntity, message, UnprocessableEntityErrorCode, err)
}

// 429 Too Many Requests - Rate limiting
func NewTooManyRequestsError(message string, err error) ApiError {
	return newApiError(http.StatusTooManyRequests, message, TooManyRequestsErrorCode, err)
}

// 500 Internal Server Error - Unexpected server error
func NewInternalServerError(message string, errCode string, err error) ApiError {
	return newApiError(http.StatusInternalServerError, message, errCode, err)
}

// 501 Not Implemented - Feature not supported by server
func NewNotImplementedError(message string, err error) ApiError {
	return newApiError(http.StatusNotImplemented, message, NotImplementedErrorCode, err)
}

// 502 Bad Gateway - Invalid response from upstream server
func NewBadGatewayError(message string, err error) ApiError {
	return newApiError(http.StatusBadGateway, message, BadGatewayErrorCode, err)
}

// 503 Service Unavailable - Server temporarily unavailable
func NewServiceUnavailableError(message string, err error) ApiError {
	return newApiError(http.StatusServiceUnavailable, message, ServiceUnavailableErrorCode, err)
}

// 504 Gateway Timeout - Upstream server timeout
func NewGatewayTimeoutError(message string, err error) ApiError {
	return newApiError(http.StatusGatewayTimeout, message, GatewayTimeoutErrorCode, err)
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func NewHTTPVersionNotSupportedError(message string, err error) ApiError {
	return newApiError(http.StatusHTTPVersionNotSupported, message, HTTPVersionNotSupportedErrorCode, err)
}
