package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ReqBody handles JSON request body binding and validation
func ReqBody[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindJSON(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest)
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())
	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity)
	}

	return dto.GetValue(), nil
}

// ReqQuery handles query parameter binding and validation
func ReqQuery[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindQuery(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest)
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())
	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity)
	}

	return dto.GetValue(), nil
}

// ReqParams handles URI parameter binding and validation
func ReqParams[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindUri(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest)
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())
	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity)
	}

	return dto.GetValue(), nil
}

// ReqHeaders handles header binding and validation
func ReqHeaders[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindHeader(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest)
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())
	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity)
	}

	return dto.GetValue(), nil
}

// handleBindingError handles binding errors (typically 400 Bad Request)
func handleBindingError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		return handleValidationError(ctx, dto, validationErrors, statusCode)
	}

	// Handle JSON syntax errors or other binding issues
	errorDetails := make([]ErrorDetail, 0)
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "JSON_SYNTAX_ERROR",
			Message: "Invalid JSON syntax",
			Detail:  syntaxErr.Error(),
		})
	} else if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "JSON_UNMARSHAL_TYPE_ERROR",
			Message: "Invalid JSON type",
			Detail:  typeErr.Error(),
		})
	} else if err.Error() == "EOF" {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "EMPTY_REQUEST_BODY",
			Message: "Empty request body",
			Detail:  "The request body is empty",
		})
	} else if err.Error() == "http: request body too large" {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "REQUEST_BODY_TOO_LARGE",
			Message: "Request body too large",
			Detail:  "The request body exceeds the maximum size limit",
		})
	} else if err.Error() == "http: request header too large" {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "REQUEST_HEADER_TOO_LARGE",
			Message: "Request header too large",
			Detail:  "The request header exceeds the maximum size limit",
		})
	} else {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "BINDING_ERROR",
			Message: "Binding error",
			Detail:  err.Error(),
		})
	}
	errResponse := NewEnvelopeWithErrors(false, statusCode, "Binding failed", errorDetails)

	ctx.JSON(statusCode, errResponse)
	return err
}

// handleValidationError handles validation errors (typically 422 Unprocessable Entity)
func handleValidationError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		msgs, e := dto.ValidateErrors(validationErrors)
		if e != nil {
			// If something went wrong during error processing
			ctx.JSON(http.StatusInternalServerError, NewEnvelopeWithErrors(
				false,
				http.StatusInternalServerError,
				"Internal server error",
				[]ErrorDetail{
					{
						Code:    InternalServerErrorCode,
						Message: "Validation error",
						Detail:  "Failed to process validation errors",
					},
				},
			))
			return e
		}

		errResponse := NewEnvelopeWithErrors(
			false,
			statusCode,
			"Validation failed",
			[]ErrorDetail{},
		)
		errResponse.Errors = make([]ErrorDetail, len(msgs))
		for i, msg := range msgs {
			errResponse.Errors[i] = ErrorDetail{
				Code:    ErrorFieldValidationCode,
				Message: msg,
				Field:   validationErrors[i].Field(),
				Detail:  fmt.Sprintf("Error: Field validation for %s failed on the %s tag", validationErrors[i].Field(), validationErrors[i].Tag()),
			}
		}
		ctx.JSON(statusCode, errResponse)

		// For compatibility with your original error return approach
		var msg strings.Builder
		br := ", "
		for _, m := range msgs {
			msg.WriteString(m + br)
		}
		// Remove the trailing separator
		errorMsg := msg.String()
		if len(errorMsg) > 0 {
			errorMsg = errorMsg[:len(errorMsg)-len(br)]
		}
		return errors.New(errorMsg)
	}

	// Handle non-validation errors
	ctx.JSON(statusCode, NewEnvelopeWithErrors(
		false,
		statusCode,
		"Validation failed",
		[]ErrorDetail{
			{
				Code:    ErrorValidationCode,
				Message: "Failed to validate request",
				Detail:  err.Error(),
			},
		}),
	)
	return err
}
