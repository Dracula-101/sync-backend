package network

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func CustomTagNameFunc() validator.TagNameFunc {
	return func(fld reflect.StructField) string {
		for _, tag := range []string{"json", "form", "uri", "param", "query"} {
			if name := strings.SplitN(fld.Tag.Get(tag), ",", 2)[0]; name != "" && name != "-" {
				return name
			}
		}
		return fld.Name
	}
}
