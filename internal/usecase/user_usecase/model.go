package user_usecase

import (
	"database/sql"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Optional menandai field PATCH: Set=false berarti field tidak dikirim.
type Optional[T any] struct {
	Set   bool
	Value T
}

// CreateInput: Role dan Division berupa kode di organisasi pemanggil.
type CreateInput struct {
	Actor       audit_usecase.Actor
	Name        string
	Email       string
	Password    string
	Role        string
	Division    sql.NullString
	UnitUsahaID sql.NullInt64
}

type UpdateInput struct {
	Actor       audit_usecase.Actor
	ID          int64
	Name        Optional[string]
	Role        Optional[string]
	Division    Optional[sql.NullString]
	UnitUsahaID Optional[sql.NullInt64]
}

type ListFilter struct {
	Q           sql.NullString
	Role        sql.NullString
	Division    sql.NullString
	UnitUsahaID sql.NullInt64
	IsActive    sql.NullBool
	Sort        string
}

type ListResult struct {
	Items []sqlc.UserDetail
	Total int64
}
