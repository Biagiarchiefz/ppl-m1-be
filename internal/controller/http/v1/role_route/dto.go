package role_route

import (
	"database/sql"
	"strings"
	"unicode/utf8"

	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/role_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/pkg/httpserver"
)

const maxNameChars = 100

type createRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

type updateRequest struct {
	Name        httpserver.Optional[string]   `json:"name"`
	Description httpserver.Optional[string]   `json:"description"`
	Permissions httpserver.Optional[[]string] `json:"permissions"`
}

type roleResponse struct {
	ID          int64    `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	IsSystem    bool     `json:"is_system"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
}

type permissionResponse struct {
	Code        string `json:"code"`
	Module      string `json:"module"`
	Description string `json:"description"`
}

func createFieldErrors(req createRequest) map[string]string {
	fields := nameErrors(&req.Name)
	if !utils.ValidCode(req.Code) {
		fields["code"] = httpserver.MsgCodeFormat
	}
	return fields
}

func updateFieldErrors(req updateRequest) map[string]string {
	fields := map[string]string{}
	if req.Name.Set {
		fields = nameErrors(req.Name.Value)
	}
	if req.Permissions.Set && req.Permissions.Value == nil {
		fields["permissions"] = httpserver.MsgFieldRequired
	}
	return fields
}

func nameErrors(name *string) map[string]string {
	fields := map[string]string{}
	switch {
	case name == nil || strings.TrimSpace(*name) == "":
		fields["name"] = httpserver.MsgFieldRequired
	case utf8.RuneCountInString(strings.TrimSpace(*name)) > maxNameChars:
		fields["name"] = "Maksimal 100 karakter"
	}
	return fields
}

func toDescription(s *string) sql.NullString {
	if s == nil || strings.TrimSpace(*s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.TrimSpace(*s), Valid: true}
}

func toUpdateInput(req updateRequest, in *role_usecase.UpdateInput) {
	if req.Name.Set {
		in.Name = req.Name.Value
	}
	if req.Description.Set {
		desc := toDescription(req.Description.Value)
		in.Description = &desc
	}
	if req.Permissions.Set {
		in.Permissions = req.Permissions.Value
	}
}

func toRoleResponse(v role_usecase.RoleView) roleResponse {
	perms := v.Permissions
	if perms == nil {
		perms = []string{}
	}
	return roleResponse{
		ID: v.Role.ID, Code: v.Role.Code, Name: v.Role.Name,
		Description: httpserver.NullStringPtr(v.Role.Description),
		IsSystem:    v.Role.IsSystem, IsActive: v.Role.IsActive,
		Permissions: perms, CreatedAt: httpserver.FormatTime(v.Role.CreatedAt),
	}
}

func toRoleList(items []role_usecase.RoleView) []roleResponse {
	out := make([]roleResponse, 0, len(items))
	for _, v := range items {
		out = append(out, toRoleResponse(v))
	}
	return out
}

func toPermissionList(items []sqlc.Permission) []permissionResponse {
	out := make([]permissionResponse, 0, len(items))
	for _, p := range items {
		out = append(out, permissionResponse{Code: p.Code, Module: p.Module, Description: p.Description})
	}
	return out
}
