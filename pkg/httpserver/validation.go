package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// UseJSONFieldNames membuat nama field di pesan validasi memakai tag json
// (snake_case), misalnya `old_password`, bukan nama field Go.
func UseJSONFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
}

func BindError(c *gin.Context, err error) {
	var verr validator.ValidationErrors
	var terr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &verr):
		FailFields(c, ErrValidation, toFieldErrors(verr))
	case errors.As(err, &terr) && terr.Field != "":
		FailFields(c, ErrValidation, map[string]string{terr.Field: MsgFieldType})
	default:
		Fail(c, ErrInvalidParameter)
	}
}

func toFieldErrors(verr validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(verr))
	for _, fe := range verr {
		fields[fe.Field()] = fieldMessage(fe)
	}
	return fields
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "email":
		return MsgFieldEmail
	case "max":
		return fmt.Sprintf("Maksimal %s karakter", fe.Param())
	}
	return MsgFieldRequired
}
