package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// ReqBody handles JSON request body binding and validation
func ReqBody[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindJSON(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest, "body")
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())

	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity, "body")
	}

	return dto.GetValue(), nil
}

// ReqForm handles form data binding and validation
func ReqForm[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	contentType := ctx.ContentType()

	var err error
	if strings.Contains(contentType, "multipart/form-data") {
		err = ctx.ShouldBindWith(dto, binding.FormMultipart)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		err = ctx.ShouldBindWith(dto, binding.Form)
	} else {
		err = ctx.ShouldBind(dto)
	}

	if err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest, "form")
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())

	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity, "form")
	}

	return dto.GetValue(), nil
}

// ReqQuery handles query parameter binding and validation
func ReqQuery[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindQuery(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest, "query")
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())

	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity, "query")
	}

	return dto.GetValue(), nil
}

// ReqParams handles URI parameter binding and validation
func ReqParams[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindUri(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest, "params")
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())

	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity, "params")
	}

	return dto.GetValue(), nil
}

// ReqHeaders handles header binding and validation
func ReqHeaders[T any](ctx *gin.Context, dto Dto[T]) (*T, error) {
	if err := ctx.ShouldBindHeader(dto); err != nil {
		return nil, handleBindingError(ctx, dto, err, http.StatusBadRequest, "header")
	}

	v := validator.New()
	v.RegisterTagNameFunc(CustomTagNameFunc())

	if err := v.Struct(dto); err != nil {
		return nil, handleValidationError(ctx, dto, err, http.StatusUnprocessableEntity, "header")
	}

	return dto.GetValue(), nil
}

// handleBindingError handles binding errors (typically 400 Bad Request)
func handleBindingError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int, bindType string) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		return handleValidationError(ctx, dto, validationErrors, statusCode, bindType)
	}

	if err == nil {
		return nil
	}

	errorDetails := make([]ErrorDetail, 0)
	errorCode := "BINDING_ERROR"
	errorMessage := fmt.Sprintf("Failed to bind %s parameters", bindType)

	switch bindType {
	case "body":
		errorCode = "BODY_BINDING_ERROR"
	case "form":
		errorCode = "FORM_BINDING_ERROR"
	case "query":
		errorCode = "QUERY_BINDING_ERROR"
	case "params":
		errorCode = "PARAMS_BINDING_ERROR"
	case "header":
		errorCode = "HEADER_BINDING_ERROR"
	}

	// Handle specific error types
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
		switch bindType {
		case "body":
			errorDetails = append(errorDetails, ErrorDetail{
				Code:    "EMPTY_BODY",
				Message: "Empty request body",
				Detail:  "The request body is empty",
			})
		case "form":
			errorDetails = append(errorDetails, ErrorDetail{
				Code:    "EMPTY_FORM",
				Message: "Empty form data",
				Detail:  "The form data is empty",
			})
		case "query":
			errorDetails = append(errorDetails, ErrorDetail{
				Code:    "EMPTY_QUERY",
				Message: "Empty query parameters",
				Detail:  "The query parameters are empty",
			})
		case "params":
			errorDetails = append(errorDetails, ErrorDetail{
				Code:    "EMPTY_PARAMS",
				Message: "Empty URI parameters",
				Detail:  "The URI parameters are empty",
			})
		case "header":
			errorDetails = append(errorDetails, ErrorDetail{
				Code:    "EMPTY_HEADER",
				Message: "Empty header parameters",
				Detail:  "The header parameters are empty",
			})
		}
	} else if err.Error() == "http: no such file" {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "FILE_NOT_FOUND",
			Message: "File not found",
			Detail:  "The specified file was not found",
		})
	} else if err.Error() == "http: request URI too large" {
		errorDetails = append(errorDetails, ErrorDetail{
			Code:    "REQUEST_URI_TOO_LARGE",
			Message: "Request URI too large",
			Detail:  "The request URI exceeds the maximum size limit",
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
			Code:    errorCode,
			Message: errorMessage,
			Detail:  err.Error(),
		})
	}
	errResponse := NewEnvelopeWithErrors(false, statusCode, "Binding failed", errorDetails)

	ctx.JSON(statusCode, errResponse)
	return err
}

// handleValidationError handles validation errors (typically 422 Unprocessable Entity)
func handleValidationError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int, bindType string) error {
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
				Field:   fmt.Sprintf("%s:%s", bindType, validationErrors[i].Field()),
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
