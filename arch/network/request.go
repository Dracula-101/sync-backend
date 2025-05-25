package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
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

// Error code and message constants
const (
	ErrCodeBindingBody     = "BODY_BINDING_ERROR"
	ErrCodeBindingForm     = "FORM_BINDING_ERROR"
	ErrCodeBindingQuery    = "QUERY_BINDING_ERROR"
	ErrCodeBindingParams   = "PARAMS_BINDING_ERROR"
	ErrCodeBindingHeader   = "HEADER_BINDING_ERROR"
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeFieldValidation = "FIELD_VALIDATION_ERROR"
	ErrCodeInternal        = "INTERNAL_SERVER_ERROR"
	ErrMsgBindingFailed    = "Failed to bind %s parameters"
	ErrMsgValidationFailed = "Validation failed"
	ErrMsgInternal         = "Internal server error"
)

// Helper: Convert validator.ValidationErrors to []ErrorDetail
func validationErrorsToDetails(validationErrors validator.ValidationErrors, bindType string, msgs []string) []ErrorDetail {
	details := make([]ErrorDetail, 0, len(validationErrors))

	// Create error details for each validation error
	for index, fieldError := range validationErrors {
		fieldName := fieldError.Field()
		field := fmt.Sprintf("%s:%s", bindType, fieldName)

		detail := NewErrorDetail(
			ErrCodeFieldValidation,
			field,
			msgs[index],
			fmt.Sprintf("Validation for '%s' failed on the '%s' tag with value '%v'",
				fieldName, fieldError.Tag(), fieldError.Value()),
			fieldError,
		)
		details = append(details, detail)
	}

	// Add any custom messages that weren't handled
	alreadyProcessedFields := make(map[string]bool)
	for _, err := range validationErrors {
		alreadyProcessedFields[err.Field()] = true
	}

	for _, msg := range msgs {
		// Check if this message is for a field we've already processed
		isProcessed := false
		for fieldName := range alreadyProcessedFields {
			if strings.Contains(strings.ToLower(msg), strings.ToLower(fieldName)) {
				isProcessed = true
				break
			}
		}

		if !isProcessed {
			detail := NewErrorDetail(
				ErrCodeFieldValidation,
				fmt.Sprintf("%s:unknown", bindType),
				msg,
				"Additional validation error",
				errors.New(msg),
			)
			details = append(details, detail)
		}
	}

	return details
}

// Helper: Extract required fields from a DTO using reflection
func getRequiredFields(dto any) []string {
	requiredFields := []string{}
	dt := reflect.TypeOf(dto)
	if dt.Kind() == reflect.Ptr {
		dt = dt.Elem()
	}
	for i := 0; i < dt.NumField(); i++ {
		field := dt.Field(i)
		// Check if field is required
		isRequired := false
		if tag, ok := field.Tag.Lookup("binding"); ok && (tag == "required" || strings.HasPrefix(tag, "required,")) {
			isRequired = true
		} else if tag, ok := field.Tag.Lookup("validate"); ok && (tag == "required" || strings.HasPrefix(tag, "required,")) {
			isRequired = true
		}

		// If required, get the proper field name from JSON tag
		if isRequired {
			fieldName := field.Name // Default to struct field name
			if jsonTag, ok := field.Tag.Lookup("json"); ok {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
				}
			}
			if jsonTag, ok := field.Tag.Lookup("form"); ok {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
				}
			}
			if jsonTag, ok := field.Tag.Lookup("query"); ok {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
				}
			}
			if jsonTag, ok := field.Tag.Lookup("uri"); ok {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
				}
			}
			requiredFields = append(requiredFields, fieldName)
		}
	}
	return requiredFields
}

// handleBindingError handles binding errors (typically 400 Bad Request)
func handleBindingError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int, bindType string) error {
	if err == nil {
		return nil
	}
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		return handleValidationError(ctx, dto, validationErrors, statusCode, bindType)
	}

	errorDetails := make([]ErrorDetail, 0)
	errorCode := map[string]string{
		"body":   ErrCodeBindingBody,
		"form":   ErrCodeBindingForm,
		"query":  ErrCodeBindingQuery,
		"params": ErrCodeBindingParams,
		"header": ErrCodeBindingHeader,
	}[bindType]
	if errorCode == "" {
		errorCode = "BINDING_ERROR"
	}
	errorMessage := fmt.Sprintf(ErrMsgBindingFailed, bindType)

	switch e := err.(type) {
	case *json.SyntaxError:
		errorDetails = append(errorDetails, NewErrorDetail(
			"JSON_SYNTAX_ERROR",
			"",
			"Invalid JSON syntax",
			fmt.Sprintf("Invalid JSON syntax at byte offset %d", e.Offset),
			e,
		))
	case *json.UnmarshalTypeError:
		errorDetails = append(errorDetails, NewErrorDetail(
			"JSON_UNMARSHAL_TYPE_ERROR",
			"",
			"Invalid JSON type",
			fmt.Sprintf("Invalid JSON type for field '%s': expected %s, got %s", e.Field, e.Type, e.Value),
			e,
		))
	default:
		switch err.Error() {
		case "EOF":
			// Empty request: enumerate all required fields
			requiredFields := getRequiredFields(dto)
			if len(requiredFields) > 0 {
				for _, field := range requiredFields {
					errorDetails = append(errorDetails, NewErrorDetail(
						ErrCodeFieldValidation,
						fmt.Sprintf("%s:%s", bindType, field),
						fmt.Sprintf("field '%s' is required", field),
						fmt.Sprintf("The field '%s' is required but was not provided", field),
						err,
					))
				}
			} else {
				errorDetails = append(errorDetails, NewErrorDetail(
					"EMPTY_"+strings.ToUpper(bindType),
					"",
					fmt.Sprintf("Empty %s", bindType),
					fmt.Sprintf("The %s is empty", bindType),
					err,
				))
			}
		case "http: no such file":
			errorDetails = append(errorDetails, NewErrorDetail(
				"FILE_NOT_FOUND", "", "File not found", "The specified file was not found", err,
			))
		case "http: request URI too large":
			errorDetails = append(errorDetails, NewErrorDetail(
				"REQUEST_URI_TOO_LARGE", "", "Request URI too large", "The request URI exceeds the maximum size limit", err,
			))
		case "http: request body too large":
			errorDetails = append(errorDetails, NewErrorDetail(
				"REQUEST_BODY_TOO_LARGE", "", "Request body too large", "The request body exceeds the maximum size limit", err,
			))
		case "http: request header too large":
			errorDetails = append(errorDetails, NewErrorDetail(
				"REQUEST_HEADER_TOO_LARGE", "", "Request header too large", "The request header exceeds the maximum size limit", err,
			))
		default:
			errorDetails = append(errorDetails, NewErrorDetail(
				errorCode, "", errorMessage, fmt.Sprintf("Failed to bind %s: %s", bindType, err.Error()), err,
			))
		}
	}

	// Fallback for truly unknown errors
	if len(errorDetails) == 0 {
		errorDetails = append(errorDetails, NewErrorDetail(
			ErrCodeInternal, "", ErrMsgInternal, fmt.Sprintf("Failed to bind %s: %s", bindType, err.Error()), err,
		))
	}

	errResponse := NewEnvelopeWithErrors(false, statusCode, "Binding failed", errorDetails)
	ctx.JSON(statusCode, errResponse)
	return err
}

// handleValidationError handles validation errors (typically 422 Unprocessable Entity)
func handleValidationError[T any](ctx *gin.Context, dto Dto[T], err error, statusCode int, bindType string) error {
	if err == nil {
		return nil
	}
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		msgs, e := dto.ValidateErrors(validationErrors)
		if e != nil {
			ctx.JSON(http.StatusInternalServerError, NewEnvelopeWithErrors(
				false,
				http.StatusInternalServerError,
				ErrMsgInternal,
				[]ErrorDetail{
					NewErrorDetail(
						ErrCodeInternal,
						"",
						"Validation error",
						fmt.Sprintf("Failed to validate request: %s", e.Error()),
						e,
					),
				},
			))
			return e
		}
		details := validationErrorsToDetails(validationErrors, bindType, msgs)
		errResponse := NewEnvelopeWithErrors(false, statusCode, ErrMsgValidationFailed, details)
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
		ErrMsgValidationFailed,
		[]ErrorDetail{
			NewErrorDetail(
				ErrCodeValidation,
				"",
				"Failed to validate request",
				fmt.Sprintf("Failed to validate %s: %s", bindType, err.Error()),
				err,
			),
		},
	))
	return err
}
