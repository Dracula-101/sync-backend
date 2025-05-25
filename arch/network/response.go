package network

import "time"

// ── Metadata envelope ─────────────────────────────────────────────────────
type Meta struct {
	Success    bool      `json:"success"`
	StatusCode int       `json:"status_code"`
	Timestamp  time.Time `json:"timestamp"`
}

// ── Envelope (generic over Data) ──────────────────────────────────────────
type Envelope struct {
	Meta    Meta          `json:"meta"`
	Message string        `json:"message,omitempty"` // optional: message
	Data    *any          `json:"data,omitempty"`    // generic data
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

// ── ErrorDetail ──────────────────────────────────────────────────────────
type ErrorDetail struct {
	Timestamp       string `json:"timestamp"`                  // RFC3339 timestamp
	Code            string `json:"code"`                       // machine‐readable error code
	Field           string `json:"field,omitempty"`            // optional: which input field
	Message         string `json:"message"`                    // human‐readable error message
	Detail          string `json:"detail,omitempty"`           // optional: developer detail
	Error           string `json:"error,omitempty"`            // optional: machine‐readable error code
	InternalMessage string `json:"internal_message,omitempty"` // debug only
	File            string `json:"file,omitempty"`             // debug only
	Function        string `json:"function,omitempty"`         // debug only
	Line            int    `json:"line,omitempty"`             // debug only
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
	if envelope.Message != "" {
		return envelope.Message
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
			Timestamp:  time.Now(),
		},
		Message: message,
		Data:    data,
	}
}

func NewEnvelopeWithErrors(success bool, statusCode int, message string, errors []ErrorDetail) Envelope {
	return Envelope{
		Meta: Meta{
			Success:    success,
			StatusCode: statusCode,
			Timestamp:  time.Now(),
		},
		Message: message,
		Errors:  errors,
	}
}

func NewErrorDetail(code string, field string, message string, detail string, err error) ErrorDetail {
	return ErrorDetail{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Code:      code,
		Field:     field,
		Message:   message,
		Detail:    detail,
		Error:     err.Error(),
	}
}

func NewErrorDetailWithDebug(code, field, message, detail, err, file, function, internalMessage string, line int) ErrorDetail {
	return ErrorDetail{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Code:            code,
		Field:           field,
		Message:         message,
		Detail:          detail,
		Error:           err,
		File:            file,
		Line:            line,
		Function:        function,
		InternalMessage: internalMessage,
	}
}
