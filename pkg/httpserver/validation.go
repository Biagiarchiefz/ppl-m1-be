package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const MsgValidation = "Data yang dikirim tidak valid"

func BindError(c *gin.Context, err error) {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		FailFields(c, MsgValidation, toFieldErrors(verr))
		return
	}
	Fail(c, http.StatusBadRequest, CodeBadRequest, "Format permintaan tidak valid")
}

func toFieldErrors(verr validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(verr))
	for _, fe := range verr {
		fields[strings.ToLower(fe.Field())] = fieldMessage(fe)
	}
	return fields
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "email":
		return "Format email tidak valid"
	case "min":
		return fmt.Sprintf("Minimal %s karakter", fe.Param())
	}
	return "Isian ini wajib diisi"
}