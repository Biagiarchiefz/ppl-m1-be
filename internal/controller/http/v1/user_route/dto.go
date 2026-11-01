package user_route

import (
	"strings"

	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/internal/usecase/utils"
)

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type registerResponse struct {
	UserID int32  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func extraFieldErrors(req registerRequest) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "Isian ini wajib diisi"
	}
	if len(req.Password) > utils.MaxPasswordBytes {
		fields["password"] = "Kata sandi terlalu panjang"
	}
	return fields
}

func toRegisterInput(req registerRequest, ip string) user_usecase.RegisterInput {
	return user_usecase.RegisterInput{
		Name: req.Name, Email: req.Email, Password: req.Password, IP: ip,
	}
}

func toRegisterResponse(r user_usecase.RegisterResult) registerResponse {
	return registerResponse{UserID: r.UserID, Name: r.Name, Email: r.Email, Role: r.Role}
}