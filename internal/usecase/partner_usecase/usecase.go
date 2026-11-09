// Package partner_usecase adalah modul M3: registrasi dan verifikasi Partner.
package partner_usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
)

const (
	StatusPending   = "pending"
	StatusVerified  = "terverifikasi"
	StatusRejected  = "ditolak"
	StatusSuspended = "ditangguhkan"
)

// transitions adalah perubahan status yang diizinkan.
var transitions = map[string][]string{
	StatusPending:   {StatusVerified, StatusRejected},
	StatusVerified:  {StatusSuspended},
	StatusSuspended: {StatusVerified},
}

type Usecase struct {
	store *repository.Store
}

func New(store *repository.Store) *Usecase {
	return &Usecase{store: store}
}

type RegisterInput struct {
	OrganizationCode string
	BusinessName     string
	BusinessType     string
	ContactName      string
	Email            string
	Password         string
	Phone            string
	Address          string
	Description      sql.NullString
	Client           audit_usecase.Client
}

type ListFilter struct {
	Status sql.NullString
	Q      sql.NullString
}

type ListResult struct {
	Items []sqlc.Partner
	Total int64
}

// Register membuat data partner berstatus pending beserta akun loginnya.
func (u *Usecase) Register(ctx context.Context, in RegisterInput) (sqlc.Partner, error) {
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return sqlc.Partner{}, err
	}
	var partner sqlc.Partner
	err = u.store.InTx(ctx, func(q *sqlc.Queries) error {
		var err error
		partner, err = register(ctx, q, in, hash)
		return err
	})
	return partner, err
}

func register(ctx context.Context, q *sqlc.Queries, in RegisterInput, hash string) (sqlc.Partner, error) {
	org, roleID, err := registrationTarget(ctx, q, in.OrganizationCode)
	if err != nil {
		return sqlc.Partner{}, err
	}
	userID, err := createAccount(ctx, q, org.ID, roleID, in, hash)
	if err != nil {
		return sqlc.Partner{}, err
	}
	partner, err := insertPartner(ctx, q, org.ID, userID, in)
	if err != nil {
		return sqlc.Partner{}, err
	}
	actor := audit_usecase.Actor{UserID: userID, OrganizationID: org.ID, Client: in.Client}
	return partner, audit_usecase.Record(ctx, q, actor.Change(org.ID, audit_usecase.ActionPartnerRegistered,
		audit_usecase.ResourcePartners, partner.ID, nil, map[string]string{"business_name": partner.BusinessName, "status": partner.Status}))
}

func registrationTarget(ctx context.Context, q *sqlc.Queries, code string) (sqlc.Organization, int64, error) {
	org, err := q.GetActiveOrganizationByCode(ctx, strings.TrimSpace(code))
	if errors.Is(err, sql.ErrNoRows) {
		return org, 0, apperr.FieldErrors{"organization_code": "Organisasi tidak ditemukan"}
	}
	if err != nil {
		return org, 0, err
	}
	roleID, err := q.GetActiveRoleIDByCode(ctx, sqlc.GetActiveRoleIDByCodeParams{OrganizationID: org.ID, Code: utils.RolePartner})
	return org, roleID, apperr.NotFound(err, apperr.PartnerRoleUnavailable)
}

func createAccount(ctx context.Context, q *sqlc.Queries, orgID, roleID int64, in RegisterInput, hash string) (int64, error) {
	id, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		OrganizationID: audit_usecase.ID(orgID), Name: strings.TrimSpace(in.ContactName),
		Email: normalizeEmail(in.Email), PasswordHash: hash, RoleID: audit_usecase.ID(roleID),
	})
	if repository.IsUniqueViolation(err) {
		return 0, apperr.EmailTaken
	}
	return id, err
}

func insertPartner(ctx context.Context, q *sqlc.Queries, orgID, userID int64, in RegisterInput) (sqlc.Partner, error) {
	return q.CreatePartner(ctx, sqlc.CreatePartnerParams{
		OrganizationID: orgID, UserID: audit_usecase.ID(userID), BusinessName: strings.TrimSpace(in.BusinessName),
		BusinessType: strings.TrimSpace(in.BusinessType), ContactName: strings.TrimSpace(in.ContactName),
		Email: normalizeEmail(in.Email), Phone: strings.TrimSpace(in.Phone), Address: strings.TrimSpace(in.Address),
		Description: in.Description,
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// OrganizationForRegistration mengenali organisasi dari tautan pendaftaran.
// Hanya organisasi aktif yang bisa menerima pendaftaran partner.
func (u *Usecase) OrganizationForRegistration(ctx context.Context, code string) (sqlc.Organization, error) {
	org, err := u.store.GetActiveOrganizationByCode(ctx, strings.TrimSpace(code))
	return org, apperr.NotFound(err, apperr.OrganizationNotFound)
}

// Organization mengambil organisasi pemanggil untuk membentuk tautan pendaftaran.
func (u *Usecase) Organization(ctx context.Context, orgID int64) (sqlc.Organization, error) {
	org, err := u.store.GetOrganization(ctx, orgID)
	return org, apperr.NotFound(err, apperr.OrganizationNotFound)
}

func (u *Usecase) List(ctx context.Context, orgID int64, f ListFilter, limit, offset int32) (ListResult, error) {
	items, err := u.store.ListPartners(ctx, sqlc.ListPartnersParams{OrganizationID: orgID, Status: f.Status, Q: f.Q, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ListResult{}, err
	}
	total, err := u.store.CountPartners(ctx, sqlc.CountPartnersParams{OrganizationID: orgID, Status: f.Status, Q: f.Q})
	return ListResult{Items: items, Total: total}, err
}

func (u *Usecase) Get(ctx context.Context, orgID, id int64) (sqlc.Partner, error) {
	partner, err := u.store.GetPartner(ctx, sqlc.GetPartnerParams{ID: id, OrganizationID: orgID})
	return partner, apperr.NotFound(err, apperr.PartnerNotFound)
}

func (u *Usecase) Mine(ctx context.Context, userID int64) (sqlc.Partner, error) {
	partner, err := u.store.GetPartnerByUser(ctx, audit_usecase.ID(userID))
	return partner, apperr.NotFound(err, apperr.PartnerNotFound)
}

func (u *Usecase) ListVerified(ctx context.Context, orgID int64) ([]sqlc.Partner, error) {
	return u.store.ListVerifiedPartners(ctx, orgID)
}
