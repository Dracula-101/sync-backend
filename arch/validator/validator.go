package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers all custom validators
func RegisterCustomValidators(v *validator.Validate) error {
	// Register custom array validator
	return v.RegisterValidation("array", ValidateArray)
}

// ValidateArray verifies if the field is a valid array
func ValidateArray(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Check if field is a slice or array type
	kind := field.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return false
	}

	// Check if the field is required to have elements
	param := fl.Param()
	if param == "nonzero" || param == "" {
		// If a non-zero length array is required, check that it has elements
		return field.Len() > 0
	}

	// For other cases, having a properly typed array/slice is sufficient
	return true
}
