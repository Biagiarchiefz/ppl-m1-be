package auth_route

import (
	"ppl-m1-be/internal/usecase/auth_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/pkg/httpserver"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// tokenUser berisi klaim token selain `exp`, untuk dipakai frontend.
type tokenUser struct {
	UserID         int64    `json:"user_id"`
	Name           string   `json:"name"`
	OrganizationID *int64   `json:"organization_id"`
	Role           string   `json:"role"`
	Permissions    []string `json:"permissions"`
	Division       *string  `json:"division"`
	UnitUsahaID    *int64   `json:"unit_usaha_id"`
}

type loginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	User        tokenUser `json:"user"`
}

// meResponse adalah kontrak K-01.
type meResponse struct {
	UserID           int64    `json:"user_id"`
	Name             string   `json:"name"`
	Email            string   `json:"email"`
	OrganizationID   *int64   `json:"organization_id"`
	OrganizationName *string  `json:"organization_name"`
	Role             string   `json:"role"`
	Permissions      []string `json:"permissions"`
	Division         *string  `json:"division"`
	UnitUsahaID      *int64   `json:"unit_usaha_id"`
	IsActive         bool     `json:"is_active"`
}

type changePasswordResponse struct {
	ChangedAt string `json:"changed_at"`
}

func toLoginResponse(r auth_usecase.LoginResult) loginResponse {
	c := r.Claims
	user := tokenUser{
		UserID: c.UserID, Name: c.Name, OrganizationID: c.OrganizationID, Role: c.Role,
		Permissions: c.Permissions, Division: c.Division, UnitUsahaID: c.UnitUsahaID,
	}
	return loginResponse{AccessToken: r.AccessToken, TokenType: "Bearer", ExpiresIn: r.ExpiresIn, User: user}
}

func toMeResponse(p auth_usecase.Profile) meResponse {
	u := p.User
	role := u.RoleCode.String
	if u.IsSuperAdmin {
		role = utils.RoleSuperAdmin
	}
	permissions := p.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	return meResponse{
		UserID: u.ID, Name: u.Name, Email: u.Email,
		OrganizationID:   httpserver.NullInt64Ptr(u.OrganizationID),
		OrganizationName: httpserver.NullStringPtr(u.OrganizationName),
		Role:             role, Permissions: permissions,
		Division:    httpserver.NullStringPtr(u.DivisionCode),
		UnitUsahaID: httpserver.NullInt64Ptr(u.UnitUsahaID),
		IsActive:    u.IsActive,
	}
}
