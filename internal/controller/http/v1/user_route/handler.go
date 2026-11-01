package user_route

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *user_usecase.Usecase
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	if fields := extraFieldErrors(req); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.MsgValidation, fields)
		return
	}

	result, err := h.uc.RegisterCustomer(c.Request.Context(), toRegisterInput(req, c.ClientIP()))
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toRegisterResponse(result))
}

func (h *Handler) writeError(c *gin.Context, err error) {
	if errors.Is(err, user_usecase.ErrEmailTaken) {
		httpserver.Fail(c, http.StatusConflict, httpserver.CodeEmailDuplicate, "Email sudah terdaftar")
		return
	}
	slog.Error("registrasi customer gagal", "error", err)
	httpserver.Fail(c, http.StatusInternalServerError, httpserver.CodeInternal, "Terjadi kesalahan pada server")
}