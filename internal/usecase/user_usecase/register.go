package user_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

func (u *Usecase) RegisterCustomer(ctx context.Context, in RegisterInput) (RegisterResult, error) {
	roleID, err := u.customerRoleID(ctx)
	if err != nil {
		return RegisterResult{}, err
	}
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return RegisterResult{}, err
	}
	user, err := u.createUser(ctx, in, hash, roleID)
	if err != nil {
		return RegisterResult{}, err
	}

	u.audit.Log(ctx, audit_usecase.Event{
		Type: "CUSTOMER_REGISTERED", UserID: user.ID, Email: user.Email, IP: in.IP,
	})
	return toRegisterResult(user), nil
}

func (u *Usecase) customerRoleID(ctx context.Context) (int32, error) {
	id, err := u.store.GetRoleIDByName(ctx, customerRoleName)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrRoleNotFound
	}
	return id, err
}

func (u *Usecase) createUser(ctx context.Context, in RegisterInput, hash string, roleID int32) (sqlc.CreateUserRow, error) {
	user, err := u.store.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(in.Name),
		RoleID:       roleID,
	})
	if isUniqueViolation(err) {
		return sqlc.CreateUserRow{}, ErrEmailTaken
	}
	return user, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toRegisterResult(u sqlc.CreateUserRow) RegisterResult {
	return RegisterResult{
		UserID: u.ID,
		Name:   u.FullName,
		Email:  u.Email,
		Role:   strings.ToLower(customerRoleName),
	}
}