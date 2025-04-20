package network

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type sender struct{}

func NewResponseSender() ResponseSender {
	return &sender{}
}

func (m *sender) Debug() bool {
	return gin.Mode() != gin.ReleaseMode
}

func (m *sender) Send(ctx *gin.Context) SendResponse {
	return &send{
		debug:   m.Debug(),
		context: ctx,
	}
}

type send struct {
	debug   bool
	context *gin.Context
}

func (s *send) SuccessMsgResponse(message string) {
	s.sendResponse(NewSuccessMsgResponse(message))
}

func (s *send) SuccessDataResponse(message string, data any) {
	s.sendResponse(NewSuccessDataResponse(message, data))
}

func (s *send) SendResourceCreatedResponse(message string) {
	s.sendResponse(NewResourceCreatedResponse(message))
}

// 400 Bad Request - Malformed request, invalid input
func (s *send) BadRequestError(message string, err error) {
	s.sendError(NewBadRequestError(message, err))
}

// 401 Unauthorized - Missing or invalid authentication
func (s *send) UnauthorizedError(message string, err error) {
	s.sendError(NewUnauthorizedError(message, err))
}

// 403 Forbidden - Valid auth but insufficient permissions
func (s *send) ForbiddenError(message string, err error) {
	s.sendError(NewForbiddenError(message, err))
}

// 404 Not Found - Resource doesn't exist
func (s *send) NotFoundError(message string, err error) {
	s.sendError(NewNotFoundError(message, err))
}

// 405 Method Not Allowed - Wrong HTTP method
func (s *send) MethodNotAllowedError(message string, err error) {
	s.sendError(NewMethodNotAllowedError(message, err))
}

// 406 Not Acceptable - Server can't fulfill requested format
func (s *send) NotAcceptableError(message string, err error) {
	s.sendError(NewNotAcceptableError(message, err))
}

// 408 Request Timeout - Client took too long to send request
func (s *send) RequestTimeoutError(message string, err error) {
	s.sendError(NewRequestTimeoutError(message, err))
}

// 409 Conflict - Request conflicts with current state
func (s *send) ConflictError(message string, err error) {
	s.sendError(NewConflictError(message, err))
}

// 410 Gone - Resource no longer available
func (s *send) ResourceGoneError(message string, err error) {
	s.sendError(NewGoneError(message, err))
}

// 413 Payload Too Large - Request entity too large
func (s *send) PayloadTooLargeError(message string, err error) {
	s.sendError(NewPayloadTooLargeError(message, err))
}

// 414 URI Too Long - Request URI too long
func (s *send) URITooLongError(message string, err error) {
	s.sendError(NewURITooLongError(message, err))
}

// 415 Unsupported Media Type - Incorrect Content-Type
func (s *send) UnsupportedMediaTypeError(message string, err error) {
	s.sendError(NewUnsupportedMediaTypeError(message, err))
}

// 419 Authentication Timeout - Custom status for expired sessions
func (s *send) SessionExpiredError(message string, err error) {
	s.sendError(NewSessionExpiredError(message, err))
}

// 422 Unprocessable Entity - Semantic errors in request
func (s *send) UnprocessableEntityError(message string, err error) {
	s.sendError(NewUnprocessableEntityError(message, err))
}

// 429 Too Many Requests - Rate limiting
func (s *send) TooManyRequestsError(message string, err error) {
	s.sendError(NewTooManyRequestsError(message, err))
}

// 500 Internal Server Error - Unexpected server error
func (s *send) InternalServerError(message string, err error) {
	s.sendError(NewInternalServerError(message, err))
}

// 501 Not Implemented - Feature not supported by server
func (s *send) NotImplementedError(message string, err error) {
	s.sendError(NewNotImplementedError(message, err))
}

// 502 Bad Gateway - Invalid response from upstream server
func (s *send) BadGatewayError(message string, err error) {
	s.sendError(NewBadGatewayError(message, err))
}

// 503 Service Unavailable - Server temporarily unavailable
func (s *send) ServiceUnavailableError(message string, err error) {
	s.sendError(NewServiceUnavailableError(message, err))
}

// 504 Gateway Timeout - Upstream server timeout
func (s *send) GatewayTimeoutError(message string, err error) {
	s.sendError(NewGatewayTimeoutError(message, err))
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func (s *send) HTTPVersionNotSupportedError(message string, err error) {
	s.sendError(NewHTTPVersionNotSupportedError(message, err))
}

func (s *send) MixedError(err error) {
	if err == nil {
		s.InternalServerError("something went wrong", err)
		return
	}

	var apiError ApiError
	if errors.As(err, &apiError) {
		s.sendError(apiError)
		return
	}

	s.InternalServerError(err.Error(), err)
}

func (s *send) sendResponse(response Response) {
	s.context.JSON(int(response.GetStatus()), response)
	// this is needed since gin calls ctx.Next() inside the resposne handeling
	// ref: https://github.com/gin-gonic/gin/issues/2221
	s.context.Abort()
}

func (s *send) sendError(err ApiError) {
	var res Response

	switch err.GetCode() {
	case http.StatusBadRequest:
		res = NewBadRequestResponse(err.GetMessage())
	case http.StatusUnauthorized:
		res = NewUnauthorizedResponse(err.GetMessage())
	case http.StatusForbidden:
		res = NewForbiddenResponse(err.GetMessage())
	case http.StatusNotFound:
		res = NewNotFoundResponse(err.GetMessage())
	case http.StatusMethodNotAllowed:
		res = NewMethodNotAllowedResponse(err.GetMessage())
	case http.StatusNotAcceptable:
		res = NewNotAcceptableResponse(err.GetMessage())
	case http.StatusRequestTimeout:
		res = NewRequestTimeoutResponse(err.GetMessage())
	case http.StatusConflict:
		res = NewConflictResponse(err.GetMessage())
	case http.StatusGone:
		res = NewGoneResponse(err.GetMessage())
	case http.StatusRequestEntityTooLarge:
		res = NewPayloadTooLargeResponse(err.GetMessage())
	case http.StatusRequestURITooLong:
		res = NewURITooLongResponse(err.GetMessage())
	case http.StatusUnsupportedMediaType:
		res = NewUnsupportedMediaTypeResponse(err.GetMessage())
	case 419: // Custom status for session expiry
		res = NewSessionExpiredResponse(err.GetMessage())
	case http.StatusUnprocessableEntity:
		res = NewUnprocessableEntityResponse(err.GetMessage())
	case http.StatusTooManyRequests:
		res = NewTooManyRequestsResponse(err.GetMessage())
	case http.StatusInternalServerError:
		res = NewInternalServerErrorResponse(err.Unwrap().Error())
	case http.StatusNotImplemented:
		res = NewNotImplementedResponse(err.GetMessage())
	case http.StatusBadGateway:
		res = NewBadGatewayResponse(err.GetMessage())
	case http.StatusServiceUnavailable:
		res = NewServiceUnavailableResponse(err.GetMessage())
	case http.StatusGatewayTimeout:
		res = NewGatewayTimeoutResponse(err.GetMessage())
	case http.StatusHTTPVersionNotSupported:
		res = NewHTTPVersionNotSupportedResponse(err.GetMessage())
	default:
		res = NewInternalServerErrorResponse(err.Unwrap().Error())
	}

	if res == nil {
		res = NewInternalServerErrorResponse("something went wrong")
	}

	s.sendResponse(res)
}
