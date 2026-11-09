package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Internal mencatat detail galat ke log server dan mengirim 500 tanpa
// membocorkan detail ke klien.
func Internal(c *gin.Context, action string, err error) {
	slog.ErrorContext(c.Request.Context(), action, "error", err)
	Fail(c, ErrInternal)
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

func SuccessList(c *gin.Context, data any, meta Meta) {
	c.JSON(200, gin.H{"success": true, "data": data, "meta": meta})
}

// Fail mengirim galat tanpa `fields` (nilainya null sesuai panduan).
func Fail(c *gin.Context, e APIError) {
	writeError(c, e, nil)
}

// FailFields mengirim galat validasi dengan `fields` berisi pesan per field.
func FailFields(c *gin.Context, e APIError, fields map[string]string) {
	writeError(c, e, fields)
}

func writeError(c *gin.Context, e APIError, fields map[string]string) {
	var payloadFields any
	if len(fields) > 0 {
		payloadFields = fields
	}
	c.JSON(e.Status, gin.H{
		"success": false,
		"error":   gin.H{"code": e.Code, "message": e.Message, "fields": payloadFields},
	})
}
