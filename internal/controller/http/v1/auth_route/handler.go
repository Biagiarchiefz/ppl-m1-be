package auth_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/auth_usecase"
	"ppl-m1-be/internal/usecase/utils"
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
	if len(req.Password) > utils.MaxPasswordBytes {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"password": httpserver.MsgPasswordTooLong})
		return
	}

	result, err := h.uc.Login(c.Request.Context(), auth_usecase.LoginInput{
		Email: req.Email, Password: req.Password, Client: middleware.Client(c),
	})
	if err != nil {
		respond.Error(c, "login gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toLoginResponse(result))
}

func (h *Handler) Me(c *gin.Context) {
	profile, err := h.uc.Me(c.Request.Context(), middleware.GetClaims(c).UserID)
	if err != nil {
		respond.Error(c, "ambil profil gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toMeResponse(profile))
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	if !utils.ValidPassword(req.NewPassword) || req.NewPassword == req.OldPassword {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"new_password": httpserver.MsgNewPasswordRule})
		return
	}

	changedAt, err := h.uc.ChangePassword(c.Request.Context(), auth_usecase.ChangePasswordInput{
		Actor: middleware.Actor(c), OldPassword: req.OldPassword, NewPassword: req.NewPassword,
	})
	if err != nil {
		respond.Error(c, "ubah kata sandi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, changePasswordResponse{ChangedAt: httpserver.FormatTime(changedAt)})
}
