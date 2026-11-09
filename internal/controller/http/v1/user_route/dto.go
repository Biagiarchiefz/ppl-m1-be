package user_route

import (
	"database/sql"
	"strings"
	"unicode/utf8"

	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/pkg/httpserver"
)

const maxNameChars = 150

type createRequest struct {
	Name        string  `json:"name" binding:"required"`
	Email       string  `json:"email" binding:"required,email,max=255"`
	Password    string  `json:"password" binding:"required"`
	Role        string  `json:"role" binding:"required"`
	Division    *string `json:"division"`
	UnitUsahaID *int64  `json:"unit_usaha_id"`
}

type updateRequest struct {
	Name        httpserver.Optional[string] `json:"name"`
	Role        httpserver.Optional[string] `json:"role"`
	Division    httpserver.Optional[string] `json:"division"`
	UnitUsahaID httpserver.Optional[int64]  `json:"unit_usaha_id"`
}

type userResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Division    *string `json:"division"`
	UnitUsahaID *int64  `json:"unit_usaha_id"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
}

type deactivateResponse struct {
	ID       int64 `json:"id"`
	IsActive bool  `json:"is_active"`
}

func createFieldErrors(req createRequest) map[string]string {
	fields := nameErrors(&req.Name)
	if !utils.ValidPassword(req.Password) {
		fields["password"] = httpserver.MsgPasswordRule
	}
	return fields
}

func updateFieldErrors(req updateRequest) map[string]string {
	fields := map[string]string{}
	if req.Name.Set {
		fields = nameErrors(req.Name.Value)
	}
	if req.Role.Set && (req.Role.Value == nil || strings.TrimSpace(*req.Role.Value) == "") {
		fields["role"] = httpserver.MsgFieldRequired
	}
	return fields
}

func nameErrors(name *string) map[string]string {
	fields := map[string]string{}
	switch {
	case name == nil || strings.TrimSpace(*name) == "":
		fields["name"] = httpserver.MsgFieldRequired
	case utf8.RuneCountInString(strings.TrimSpace(*name)) > maxNameChars:
		fields["name"] = "Maksimal 150 karakter"
	}
	return fields
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullInt64(n *int64) sql.NullInt64 {
	if n == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *n, Valid: true}
}

func toUpdateFields(req updateRequest, in *user_usecase.UpdateInput) {
	if req.Name.Set {
		in.Name = user_usecase.Optional[string]{Set: true, Value: *req.Name.Value}
	}
	if req.Role.Set {
		in.Role = user_usecase.Optional[string]{Set: true, Value: *req.Role.Value}
	}
	if req.Division.Set {
		in.Division = user_usecase.Optional[sql.NullString]{Set: true, Value: toNullString(req.Division.Value)}
	}
	if req.UnitUsahaID.Set {
		in.UnitUsahaID = user_usecase.Optional[sql.NullInt64]{Set: true, Value: toNullInt64(req.UnitUsahaID.Value)}
	}
}

func toUserResponse(u sqlc.UserDetail) userResponse {
	return userResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Role: u.RoleCode.String,
		Division:    httpserver.NullStringPtr(u.DivisionCode),
		UnitUsahaID: httpserver.NullInt64Ptr(u.UnitUsahaID),
		IsActive:    u.IsActive,
		CreatedAt:   httpserver.FormatTime(u.CreatedAt),
	}
}

func toUserList(users []sqlc.UserDetail) []userResponse {
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toUserResponse(u))
	}
	return out
}
