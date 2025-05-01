package network

import (
	"time"
)

// ── Metadata envelope ─────────────────────────────────────────────────────
type Meta struct {
	Success    bool      `json:"success"`
	StatusCode int       `json:"status_code"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// ── Envelope (generic over Data) ──────────────────────────────────────────
type Envelope struct {
	Meta   Meta          `json:"meta"`
	Data   *any          `json:"data,omitempty"` // generic data
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// ── ErrorDetail ──────────────────────────────────────────────────────────
type ErrorDetail struct {
	Code    string `json:"code"`             // machine‐readable error code
	Field   string `json:"field,omitempty"`  // optional: which input field
	Message string `json:"message"`          // user‐friendly message
	Detail  string `json:"detail,omitempty"` // optional: developer detail
}

func (envelope Envelope) GetStatusCode() int {
	if envelope.Meta.StatusCode != 0 {
		return envelope.Meta.StatusCode
	}
	if envelope.Meta.Success {
		return 200
	}
	return 500
}

func (envelope Envelope) GetStatus() bool {
	return envelope.Meta.Success
}

func (envelope Envelope) GetMessage() string {
	if envelope.Meta.Message != "" {
		return envelope.Meta.Message
	}
	if envelope.Meta.Success {
		return "Success"
	}
	return "Error"
}

func (envelope Envelope) GetData() *any {
	if envelope.Data != nil {
		return envelope.Data
	}
	return nil
}

func (envelope Envelope) GetErrors() *[]ErrorDetail {
	if envelope.Errors != nil {
		return &envelope.Errors
	}
	return nil
}

func NewEnvelopeWithData(success bool, statusCode int, message string, data *any) Envelope {
	return Envelope{
		Meta: Meta{
			Success:    success,
			StatusCode: statusCode,
			Message:    message,
			Timestamp:  time.Now(),
		},
		Data: data,
	}
}

func NewEnvelopeWithErrors(success bool, statusCode int, message string, errors []ErrorDetail) Envelope {
	return Envelope{
		Meta: Meta{
			Success:    success,
			StatusCode: statusCode,
			Message:    message,
			Timestamp:  time.Now(),
		},
		Errors: errors,
	}
}

func NewErrorDetail(code string, field string, message string, detail string) ErrorDetail {
	return ErrorDetail{
		Code:    code,
		Field:   field,
		Message: message,
		Detail:  detail,
	}
}
