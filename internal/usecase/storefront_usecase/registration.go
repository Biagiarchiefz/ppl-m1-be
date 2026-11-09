package storefront_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/transaction_usecase"
	"ppl-m1-be/internal/usecase/utils"
)

type CustomerRegistration struct {
	OrganizationCode string
	Name             string
	Email            string
	Password         string
	Phone            string
	Client           audit_usecase.Client
}

// OrganizationForRegistration mengenali organisasi dari tautan pendaftaran.
func (u *Usecase) OrganizationForRegistration(ctx context.Context, code string) (sqlc.Organization, error) {
	org, err := u.store.GetActiveOrganizationByCode(ctx, strings.TrimSpace(code))
	return org, apperr.NotFound(err, apperr.OrganizationNotFound)
}

// RegisterCustomer membuat akun customer di organisasi tujuan. Data pelanggan
// dengan email yang sama yang sudah diinput penjual ikut ditautkan.
func (u *Usecase) RegisterCustomer(ctx context.Context, in CustomerRegistration) error {
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return err
	}
	return u.store.InTx(ctx, func(q *sqlc.Queries) error {
		orgID, userID, err := createCustomerAccount(ctx, q, in, hash)
		if err != nil {
			return err
		}
		customerID, err := linkCustomer(ctx, q, userID, in.Phone)
		if err != nil {
			return err
		}
		actor := audit_usecase.Actor{UserID: userID, OrganizationID: orgID, Client: in.Client}
		return audit_usecase.Record(ctx, q, actor.Change(orgID, audit_usecase.ActionCustomerRegistered,
			audit_usecase.ResourceCustomers, customerID, nil, map[string]string{"email": strings.ToLower(strings.TrimSpace(in.Email))}))
	})
}

func createCustomerAccount(ctx context.Context, q *sqlc.Queries, in CustomerRegistration, hash string) (int64, int64, error) {
	org, err := q.GetActiveOrganizationByCode(ctx, strings.TrimSpace(in.OrganizationCode))
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, apperr.FieldErrors{"organization_code": "Organisasi tidak ditemukan"}
	}
	if err != nil {
		return 0, 0, err
	}
	roleID, err := q.GetActiveRoleIDByCode(ctx, sqlc.GetActiveRoleIDByCodeParams{OrganizationID: org.ID, Code: utils.RoleCustomer})
	if err != nil {
		return 0, 0, apperr.NotFound(err, apperr.CustomerRoleUnavailable)
	}
	userID, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		OrganizationID: audit_usecase.ID(org.ID), Name: strings.TrimSpace(in.Name),
		Email: strings.ToLower(strings.TrimSpace(in.Email)), PasswordHash: hash, RoleID: audit_usecase.ID(roleID),
	})
	if repository.IsUniqueViolation(err) {
		return 0, 0, apperr.EmailTaken
	}
	return org.ID, userID, err
}

func linkCustomer(ctx context.Context, q *sqlc.Queries, userID int64, phone string) (int64, error) {
	customerID, err := transaction_usecase.EnsureCustomerForUser(ctx, q, userID)
	if err != nil {
		return 0, err
	}
	return customerID, q.SetCustomerPhoneIfEmpty(ctx, sqlc.SetCustomerPhoneIfEmptyParams{
		ID: customerID, Phone: sql.NullString{String: strings.TrimSpace(phone), Valid: true},
	})
}
