package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeBadRequest         = "BAD_REQUEST"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeForbidden          = "FORBIDDEN"
	CodeValidation         = "VALIDATION_ERROR"
	CodeInternal           = "INTERNAL_ERROR"
	CodeEmailDuplicate     = "EMAIL_DUPLICATE"
)

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   gin.H{"code": code, "message": message},
	})
}

func FailFields(c *gin.Context, message string, fields map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"error": gin.H{
			"code":    CodeValidation,
			"message": message,
			"fields":  fields,
		},
	})
}