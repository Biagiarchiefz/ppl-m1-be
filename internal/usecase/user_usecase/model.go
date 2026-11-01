package user_usecase

import (
	"context"
	"errors"

	"ppl-m1-be/internal/usecase/repository/sqlc"
)

const customerRoleName = "customer"

var (
	ErrEmailTaken   = errors.New("email sudah terdaftar")
	ErrRoleNotFound = errors.New("role customer belum tersedia di tabel roles")
)

type userStore interface {
	GetRoleIDByName(ctx context.Context, roleName string) (int32, error)
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.CreateUserRow, error)
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	IP       string
}

type RegisterResult struct {
	UserID int32
	Name   string
	Email  string
	Role   string
}