package network

import (
	"errors"

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
	res := NewEnvelopeWithData(true, 200, message, nil)
	s.sendResponse(res)
}

func (s *send) SuccessDataResponse(message string, data any) {
	res := NewEnvelopeWithData(true, 200, message, &data)
	s.sendResponse(res)
}

// 400 Bad Request - Malformed request, invalid input
func (s *send) BadRequestError(message string, detail string, err error) {
	s.sendError(NewBadRequestError(message, detail, err))
}

// 401 Unauthorized - Missing or invalid authentication
func (s *send) UnauthorizedError(message string, detail string, err error) {
	s.sendError(NewUnauthorizedError(message, detail, err))
}

// 403 Forbidden - Valid auth but insufficient permissions
func (s *send) ForbiddenError(message string, detail string, err error) {
	s.sendError(NewForbiddenError(message, detail, err))
}

// 404 Not Found - Resource doesn't exist
func (s *send) NotFoundError(message string, detail string, err error) {
	s.sendError(NewNotFoundError(message, detail, err))
}

// 405 Method Not Allowed - Wrong HTTP method
func (s *send) MethodNotAllowedError(message string, detail string, err error) {
	s.sendError(NewMethodNotAllowedError(message, detail, err))
}

// 406 Not Acceptable - Server can't fulfill requested format
func (s *send) NotAcceptableError(message string, detail string, err error) {
	s.sendError(NewNotAcceptableError(message, detail, err))
}

// 408 Request Timeout - Client took too long to send request
func (s *send) RequestTimeoutError(message string, detail string, err error) {
	s.sendError(NewRequestTimeoutError(message, detail, err))
}

// 409 Conflict - Request conflicts with current state
func (s *send) ConflictError(message string, detail string, err error) {
	s.sendError(NewConflictError(message, detail, err))
}

// 410 Gone - Resource no longer available
func (s *send) ResourceGoneError(message string, detail string, err error) {
	s.sendError(NewGoneError(message, detail, err))
}

// 413 Payload Too Large - Request entity too large
func (s *send) PayloadTooLargeError(message string, detail string, err error) {
	s.sendError(NewPayloadTooLargeError(message, detail, err))
}

// 414 URI Too Long - Request URI too long
func (s *send) URITooLongError(message string, detail string, err error) {
	s.sendError(NewURITooLongError(message, detail, err))
}

// 415 Unsupported Media Type - Incorrect Content-Type
func (s *send) UnsupportedMediaTypeError(message string, detail string, err error) {
	s.sendError(NewUnsupportedMediaTypeError(message, detail, err))
}

// 419 Authentication Timeout - Custom status for expired sessions
func (s *send) SessionExpiredError(message string, detail string, err error) {
	s.sendError(NewSessionExpiredError(message, detail, err))
}

// 422 Unprocessable Entity - Semantic errors in request
func (s *send) UnprocessableEntityError(message string, detail string, err error) {
	s.sendError(NewUnprocessableEntityError(message, detail, err))
}

// 429 Too Many Requests - Rate limiting
func (s *send) TooManyRequestsError(message string, detail string, err error) {
	s.sendError(NewTooManyRequestsError(message, detail, err))
}

// 500 Internal Server Error - Unexpected server error
func (s *send) InternalServerError(message string, detail string, errCode string, err error) {
	s.sendError(NewInternalServerError(message, errCode, detail, err))
}

// 501 Not Implemented - Feature not supported by server
func (s *send) NotImplementedError(message string, detail string, err error) {
	s.sendError(NewNotImplementedError(message, detail, err))
}

// 502 Bad Gateway - Invalid response from upstream server
func (s *send) BadGatewayError(message string, detail string, err error) {
	s.sendError(NewBadGatewayError(message, detail, err))
}

// 503 Service Unavailable - Server temporarily unavailable
func (s *send) ServiceUnavailableError(message string, detail string, err error) {
	s.sendError(NewServiceUnavailableError(message, detail, err))
}

// 504 Gateway Timeout - Upstream server timeout
func (s *send) GatewayTimeoutError(message string, detail string, err error) {
	s.sendError(NewGatewayTimeoutError(message, detail, err))
}

// 505 HTTP Version Not Supported - Unsupported HTTP version
func (s *send) HTTPVersionNotSupportedError(message string, detail string, err error) {
	s.sendError(NewHTTPVersionNotSupportedError(message, detail, err))
}

func (s *send) MixedError(err error) {
	if err == nil {
		s.InternalServerError("Something went wrong", "Server encountered something unexpected and failed to process the request", UnknownErrorCode, nil)
		return
	}

	var apiError ApiError
	if errors.As(err, &apiError) {
		s.sendError(apiError)
		return
	}

	s.InternalServerError(err.Error(), "Server encountered something unexpected and failed to process the request", UnknownErrorCode, err)
}

func (s *send) sendResponse(response Response) {
	s.context.JSON(response.GetStatusCode(), response)
	// this is needed since gin calls ctx.Next() inside the resposne handeling
	// ref: https://github.com/gin-gonic/gin/issues/2221
	s.context.Abort()
}

func (s *send) sendError(err ApiError) {
	res := NewEnvelopeWithErrors(false, err.GetStatusCode(), err.GetMessage(), err.GetErrors(s.debug))
	s.sendResponse(res)
}
