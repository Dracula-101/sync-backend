package coredto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func EmptyPagination() *Pagination {
	return &Pagination{
		Page:  1,
		Limit: 10,
	}
}

type Pagination struct {
	Page  int `form:"page" query:"page" validate:"min=1"`
	Limit int `form:"limit" query:"limit" validate:"min=1,max=100"`
}

func (d *Pagination) GetValue() *Pagination {
	return d
}

func (d *Pagination) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be min %s", err.Field(), err.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be max%s", err.Field(), err.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}
