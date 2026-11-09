package auth_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

// Login memeriksa kredensial. ACCOUNT_INACTIVE dan ORGANIZATION_INACTIVE
// hanya dikembalikan setelah kata sandi terbukti benar (PRD 10.2).
func (u *Usecase) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	email := strings.TrimSpace(in.Email)
	user, found, err := u.findUser(ctx, email)
	if err != nil {
		return LoginResult{}, err
	}

	if reason := u.failReason(user, found, in.Password); reason != "" {
		if err := u.recordLoginFailed(ctx, user, email, reason, in.Client); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{}, reasonError(reason)
	}
	return u.issueToken(ctx, user, in.Client)
}

func (u *Usecase) findUser(ctx context.Context, email string) (sqlc.UserDetail, bool, error) {
	user, err := u.store.GetUserDetailByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return sqlc.UserDetail{}, false, nil
	}
	return user, err == nil, err
}

func (u *Usecase) failReason(user sqlc.UserDetail, found bool, password string) string {
	hash := user.PasswordHash
	if !found {
		hash = u.dummyHash
	}
	wrongPassword := utils.CheckPassword(hash, password) != nil

	switch {
	case !found:
		return reasonUnknownEmail
	case wrongPassword:
		return reasonWrongPassword
	}
	return accessError(user)
}

func reasonError(reason string) error {
	switch reason {
	case reasonInactive:
		return apperr.AccountInactive
	case reasonOrgInactive:
		return apperr.OrganizationInactive
	}
	return apperr.InvalidCredentials
}

func (u *Usecase) recordLoginFailed(ctx context.Context, user sqlc.UserDetail, email, reason string, client audit_usecase.Client) error {
	return audit_usecase.Record(ctx, u.store.Queries, audit_usecase.Entry{
		OrganizationID: user.OrganizationID,
		ActorID:        audit_usecase.ID(user.ID),
		Action:         audit_usecase.ActionLoginFailed,
		Resource:       audit_usecase.ResourceAuth,
		ResourceID:     audit_usecase.ID(user.ID),
		NewValue:       map[string]string{"email": email, "reason": reason},
		Client:         client,
	})
}

func (u *Usecase) issueToken(ctx context.Context, user sqlc.UserDetail, client audit_usecase.Client) (LoginResult, error) {
	permissions, err := permissionsOf(ctx, u.store.Queries, user)
	if err != nil {
		return LoginResult{}, err
	}
	claims := toClaims(user, permissions)
	token, err := u.tokens.Generate(claims)
	if err != nil {
		return LoginResult{}, err
	}
	err = audit_usecase.Record(ctx, u.store.Queries, audit_usecase.Entry{
		OrganizationID: user.OrganizationID, ActorID: audit_usecase.ID(user.ID),
		Action: audit_usecase.ActionLoginSucceeded, Resource: audit_usecase.ResourceAuth,
		ResourceID: audit_usecase.ID(user.ID), Client: client,
	})
	return LoginResult{AccessToken: token, ExpiresIn: int(u.tokens.TTL().Seconds()), Claims: claims}, err
}
