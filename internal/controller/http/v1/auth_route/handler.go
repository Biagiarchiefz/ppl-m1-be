package auth_route

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/auth_usecase"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *auth_usecase.Usecase
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	if len(req.Password) > maxPasswordBytes {
		httpserver.FailFields(c, httpserver.MsgValidation, map[string]string{"password": "Kata sandi terlalu panjang"})
		return
	}

	result, err := h.uc.Login(c.Request.Context(), toLoginInput(req, c.ClientIP()))
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpserver.Success(c, http.StatusOK, toLoginResponse(result))
}

func (h *Handler) writeError(c *gin.Context, err error) {
	if errors.Is(err, auth_usecase.ErrInvalidCredentials) {
		httpserver.Fail(c, http.StatusUnauthorized, httpserver.CodeInvalidCredentials, "Email atau kata sandi salah")
		return
	}
	slog.Error("login gagal", "error", err)
	httpserver.Fail(c, http.StatusInternalServerError, httpserver.CodeInternal, "Terjadi kesalahan pada server")
}

func (h *Handler) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	httpserver.Success(c, http.StatusOK, toMeResponse(claims))
}

func (h *Handler) Logout(c *gin.Context) {
	httpserver.Success(c, http.StatusOK, nil)
}