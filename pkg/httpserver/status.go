package httpserver

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

// StatusUpdate membaca permintaan PATCH /{resource}/{id}/status: id dari
// alamat dan is_active dari body. Bila gagal, responsnya sudah dikirim dan
// ok bernilai false.
func StatusUpdate(c *gin.Context) (ctx context.Context, id int64, active, ok bool) {
	id, ok = PathID(c)
	if !ok {
		Fail(c, ErrInvalidParameter)
		return nil, 0, false, false
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BindError(c, err)
		return nil, 0, false, false
	}
	return c.Request.Context(), id, *req.IsActive, true
}

// StatusUpdated mengirim hasil perubahan status.
func StatusUpdated(c *gin.Context, id int64, active bool) {
	Success(c, http.StatusOK, gin.H{"id": id, "is_active": active})
}
