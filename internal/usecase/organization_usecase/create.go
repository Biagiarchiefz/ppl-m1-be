package organization_usecase

import (
	"context"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

type AdminInput struct {
	Name     string
	Email    string
	Password string
}

type CreateInput struct {
	Actor audit_usecase.Actor
	Code  string
	Name  string
	Admin AdminInput
}

type CreateResult struct {
	Organization sqlc.Organization
	AdminID      int64
	AdminEmail   string
}

// Create mendaftarkan organisasi baru beserta peran administrator bawaan
// dan akun administrator pertamanya, dalam satu transaksi.
func (u *Usecase) Create(ctx context.Context, in CreateInput) (CreateResult, error) {
	hash, err := utils.HashPassword(in.Admin.Password)
	if err != nil {
		return CreateResult{}, err
	}
	var res CreateResult
	err = u.store.InTx(ctx, func(q *sqlc.Queries) error {
		org, err := q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{Code: in.Code, Name: strings.TrimSpace(in.Name)})
		if repository.IsUniqueViolation(err) {
			return apperr.OrganizationCodeTaken
		}
		if err != nil {
			return err
		}
		res, err = createAdmin(ctx, q, org, in, hash)
		if err != nil {
			return err
		}
		return recordCreated(ctx, q, in.Actor, res)
	})
	return res, err
}

func createAdmin(ctx context.Context, q *sqlc.Queries, org sqlc.Organization, in CreateInput, hash string) (CreateResult, error) {
	role, err := q.CreateRole(ctx, sqlc.CreateRoleParams{
		OrganizationID: org.ID, Code: utils.SystemRoleCode, Name: utils.SystemRoleName, IsSystem: true,
	})
	if err != nil {
		return CreateResult{}, err
	}
	email := strings.ToLower(strings.TrimSpace(in.Admin.Email))
	id, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		OrganizationID: audit_usecase.ID(org.ID), Name: strings.TrimSpace(in.Admin.Name), Email: email,
		PasswordHash: hash, RoleID: audit_usecase.ID(role.ID), CreatedBy: audit_usecase.ID(in.Actor.UserID),
	})
	if repository.IsUniqueViolation(err) {
		return CreateResult{}, apperr.EmailTaken
	}
	return CreateResult{Organization: org, AdminID: id, AdminEmail: email}, err
}

func recordCreated(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, res CreateResult) error {
	org := res.Organization
	err := audit_usecase.Record(ctx, q, actor.Change(org.ID, audit_usecase.ActionOrgCreated,
		audit_usecase.ResourceOrganizations, org.ID, nil, map[string]string{"code": org.Code, "name": org.Name}))
	if err != nil {
		return err
	}
	return audit_usecase.Record(ctx, q, actor.Change(org.ID, audit_usecase.ActionUserCreated,
		audit_usecase.ResourceUsers, res.AdminID, nil, map[string]string{"email": res.AdminEmail, "role": utils.SystemRoleCode}))
}
