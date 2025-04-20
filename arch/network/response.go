package network

import (
	"net/http"
)

type response struct {
	Status  int    `json:"status" binding:"required"`
	Message string `json:"message" binding:"required"`
	Data    any    `json:"data,omitempty" binding:"required,omitempty"`
}

func (r *response) GetStatus() int {
	return r.Status
}

func (r *response) GetMessage() string {
	return r.Message
}

func (r *response) GetData() any {
	return r.Data
}

// 200 OK - Request succeeded
func NewSuccessDataResponse(message string, data any) Response {
	return &response{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	}
}

// 200 OK - Request succeeded with message only
func NewSuccessMsgResponse(message string) Response {
	return &response{
		Status:  http.StatusOK,
		Message: message,
	}
}

// 201 Created - Resource successfully created
func NewResourceCreatedResponse(message string) Response {
	return &response{
		Status:  http.StatusCreated,
		Message: message,
	}
}

// 400 Bad Request - Malformed request, invalid input
func NewBadRequestResponse(message string) Response {
	return &response{
		Status:  http.StatusBadRequest,
		Message: message,
	}
}

// 401 Unauthorized - Missing or invalid authentication
func NewUnauthorizedResponse(message string) Response {
	return &response{
		Status:  http.StatusUnauthorized,
		Message: message,
	}
}

// 403 Forbidden - Valid auth but insufficient permissions
func NewForbiddenResponse(message string) Response {
	return &response{
		Status:  http.StatusForbidden,
		Message: message,
	}
}

// 404 Not Found - Resource doesn't exist
func NewNotFoundResponse(message string) Response {
	return &response{
		Status:  http.StatusNotFound,
		Message: message,
	}
}

// 405 Method Not Allowed - Wrong HTTP method
func NewMethodNotAllowedResponse(message string) Response {
	return &response{
		Status:  http.StatusMethodNotAllowed,
		Message: message,
	}
}

// 406 Not Acceptable - Server can't fulfill requested format
func NewNotAcceptableResponse(message string) Response {
	return &response{
		Status:  http.StatusNotAcceptable,
		Message: message,
	}
}

// 408 Request Timeout - Client took too long to send request
func NewRequestTimeoutResponse(message string) Response {
	return &response{
		Status:  http.StatusRequestTimeout,
		Message: message,
	}
}

// 409 Conflict - Request conflicts with current state
func NewConflictResponse(message string) Response {
	return &response{
		Status:  http.StatusConflict,
		Message: message,
	}
}

// 410 Gone - Resource no longer available
func NewGoneResponse(message string) Response {
	return &response{
		Status:  http.StatusGone,
		Message: message,
	}
}

// 413 Payload Too Large - Request entity too large
func NewPayloadTooLargeResponse(message string) Response {
	return &response{
		Status:  http.StatusRequestEntityTooLarge,
		Message: message,
	}
}

// 414 URI Too Long - Request URI too long
func NewURITooLongResponse(message string) Response {
	return &response{
		Status:  http.StatusRequestURITooLong,
		Message: message,
	}
}

// 415 Unsupported Media Type - Incorrect Content-Type
func NewUnsupportedMediaTypeResponse(message string) Response {
	return &response{
		Status:  http.StatusUnsupportedMediaType,
		Message: message,
	}
}

// 419 Authentication Timeout - Custom status for expired sessions
func NewSessionExpiredResponse(message string) Response {
	return &response{
		Status:  419,
		Message: message,
	}
}

// 422 Unprocessable Entity - Semantic errors in request
func NewUnprocessableEntityResponse(message string) Response {
	return &response{
		Status:  http.StatusUnprocessableEntity,
		Message: message,
	}
}

// 429 Too Many Requests - Rate limiting
func NewTooManyRequestsResponse(message string) Response {
	return &response{
		Status:  http.StatusTooManyRequests,
		Message: message,
	}
}

// 500 Internal Server Error - Unexpected server error
func NewInternalServerErrorResponse(message string) Response {
	return &response{
		Status:  http.StatusInternalServerError,
		Message: message,
	}
}

// 501 Not Implemented - Feature not supported by server
func NewNotImplementedResponse(message string) Response {
	return &response{
		Status:  http.StatusNotImplemented,
		Message: message,
	}
}

// 502 Bad Gateway - Invalid response from upstream server
func NewBadGatewayResponse(message string) Response {
	return &response{
		Status:  http.StatusBadGateway,
		Message: message,
	}
}

// 503 Service Unavailable - Server temporarily unavailable
func NewServiceUnavailableResponse(message string) Response {
	return &response{
		Status:  http.StatusServiceUnavailable,
		Message: message,
	}
}

// 504 Gateway Timeout - Upstream server timeout
func NewGatewayTimeoutResponse(message string) Response {
	return &response{
		Status:  http.StatusGatewayTimeout,
		Message: message,
	}
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func NewHTTPVersionNotSupportedResponse(message string) Response {
	return &response{
		Status:  http.StatusHTTPVersionNotSupported,
		Message: message,
	}
}
