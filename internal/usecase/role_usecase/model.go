package role_usecase

import (
	"database/sql"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// RoleView adalah peran beserta kode izinnya.
type RoleView struct {
	Role        sqlc.Role
	Permissions []string
}

type CreateInput struct {
	Actor       audit_usecase.Actor
	Code        string
	Name        string
	Description sql.NullString
	Permissions []string
}

// UpdateInput: nil berarti field tidak dikirim (tidak diubah).
type UpdateInput struct {
	Actor       audit_usecase.Actor
	ID          int64
	Name        *string
	Description *sql.NullString
	Permissions *[]string
}

type ListResult struct {
	Items []RoleView
	Total int64
}
