package network

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	StatusCode int      `json:"status_code"`
	Message    string   `json:"message"`
	Details    []string `json:"details,omitempty"`
}

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
	errResponse := ErrorResponse{
		StatusCode: statusCode,
		Message:    "Invalid request format",
	}

	ctx.JSON(statusCode, errResponse)
	return errors.New(errResponse.Message)
}

// handleValidationError handles validation errors (typically 422 Unprocessable Entity)
func handleValidationError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		msgs, e := dto.ValidateErrors(validationErrors)
		if e != nil {
			// If something went wrong during error processing
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "Error processing validation",
			})
			return e
		}

		errResponse := ErrorResponse{
			StatusCode: statusCode,
			Message:    "Validation failed",
			Details:    msgs,
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
	ctx.JSON(statusCode, ErrorResponse{
		StatusCode: statusCode,
		Message:    err.Error(),
	})
	return err
}
