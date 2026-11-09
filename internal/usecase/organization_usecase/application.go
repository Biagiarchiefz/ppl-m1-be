package organization_usecase

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
	ApplicationPending  = "pending"
	ApplicationApproved = "disetujui"
	ApplicationRejected = "ditolak"
)

type ApplicationInput struct {
	Code             string
	OrganizationName string
	AdminName        string
	AdminEmail       string
	Password         string
	Phone            string
	Description      sql.NullString
	Client           audit_usecase.Client
}

type ApplicationPage struct {
	Items []sqlc.OrganizationApplication
	Total int64
}

// Apply menyimpan pengajuan berstatus pending. Organisasi dan akunnya belum
// dibuat sampai super admin menyetujui.
func (u *Usecase) Apply(ctx context.Context, in ApplicationInput) (sqlc.OrganizationApplication, error) {
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return sqlc.OrganizationApplication{}, err
	}
	var application sqlc.OrganizationApplication
	err = u.store.InTx(ctx, func(q *sqlc.Queries) error {
		if err := checkApplicationFree(ctx, q, in); err != nil {
			return err
		}
		var err error
		if application, err = insertApplication(ctx, q, in, hash); err != nil {
			return err
		}
		return audit_usecase.Record(ctx, q, audit_usecase.Entry{
			Action: audit_usecase.ActionApplicationSubmitted, Resource: audit_usecase.ResourceApplications, ResourceID: audit_usecase.ID(application.ID),
			NewValue: map[string]string{"code": application.Code, "name": application.OrganizationName}, Client: in.Client,
		})
	})
	return application, err
}

func checkApplicationFree(ctx context.Context, q *sqlc.Queries, in ApplicationInput) error {
	taken, err := q.OrganizationCodeExists(ctx, in.Code)
	if err != nil {
		return err
	}
	if taken {
		return apperr.OrganizationCodeTaken
	}
	_, err = q.GetUserDetailByEmail(ctx, normalizeEmail(in.AdminEmail))
	if err == nil {
		return apperr.EmailTaken
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func insertApplication(ctx context.Context, q *sqlc.Queries, in ApplicationInput, hash string) (sqlc.OrganizationApplication, error) {
	application, err := q.CreateOrganizationApplication(ctx, sqlc.CreateOrganizationApplicationParams{
		Code: in.Code, OrganizationName: strings.TrimSpace(in.OrganizationName), AdminName: strings.TrimSpace(in.AdminName),
		AdminEmail: normalizeEmail(in.AdminEmail), PasswordHash: hash, Phone: strings.TrimSpace(in.Phone), Description: in.Description,
	})
	if repository.IsUniqueViolation(err) {
		return application, apperr.ApplicationPending
	}
	return application, err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ApplicationStatus dipakai pemohon yang belum punya akun. Kata sandi
// pengajuan menjadi bukti bahwa pemohon berhak melihat statusnya.
func (u *Usecase) ApplicationStatus(ctx context.Context, email, password string) (sqlc.OrganizationApplication, error) {
	application, err := u.store.GetLatestOrganizationApplication(ctx, normalizeEmail(email))
	if errors.Is(err, sql.ErrNoRows) {
		return application, apperr.InvalidCredentials
	}
	if err != nil {
		return application, err
	}
	if utils.CheckPassword(application.PasswordHash, password) != nil {
		return sqlc.OrganizationApplication{}, apperr.InvalidCredentials
	}
	return application, nil
}

func (u *Usecase) ListApplications(ctx context.Context, status sql.NullString, limit, offset int32) (ApplicationPage, error) {
	items, err := u.store.ListOrganizationApplications(ctx, sqlc.ListOrganizationApplicationsParams{Status: status, LimitRows: limit, OffsetRows: offset})
	if err != nil {
		return ApplicationPage{}, err
	}
	total, err := u.store.CountOrganizationApplications(ctx, status)
	return ApplicationPage{Items: items, Total: total}, err
}

// ReviewApplication menyetujui atau menolak pengajuan yang masih menunggu.
// Persetujuan membuat organisasi dan akun administrator dalam satu transaksi.
func (u *Usecase) ReviewApplication(ctx context.Context, actor audit_usecase.Actor, id int64, status, reason string) (sqlc.OrganizationApplication, error) {
	reason = strings.TrimSpace(reason)
	if status == ApplicationRejected && reason == "" {
		return sqlc.OrganizationApplication{}, apperr.FieldErrors{"reason": "Alasan wajib diisi"}
	}
	var reviewed sqlc.OrganizationApplication
	err := u.store.InTx(ctx, func(q *sqlc.Queries) error {
		var err error
		reviewed, err = reviewApplication(ctx, q, actor, id, status, reason)
		return err
	})
	return reviewed, err
}

func reviewApplication(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, id int64, status, reason string) (sqlc.OrganizationApplication, error) {
	current, err := lockPendingApplication(ctx, q, id)
	if err != nil {
		return current, err
	}
	var orgID int64
	if status == ApplicationApproved {
		if orgID, err = createFromApplication(ctx, q, actor, current); err != nil {
			return current, err
		}
	}
	reviewed, err := q.ReviewOrganizationApplication(ctx, sqlc.ReviewOrganizationApplicationParams{
		ID: id, Status: status, StatusReason: sql.NullString{String: reason, Valid: reason != ""},
		OrganizationID: audit_usecase.ID(orgID), ReviewedBy: audit_usecase.ID(actor.UserID),
	})
	if err != nil {
		return reviewed, err
	}
	return reviewed, audit_usecase.Record(ctx, q, actor.Change(orgID, audit_usecase.ActionApplicationReviewed,
		audit_usecase.ResourceApplications, id, map[string]string{"status": current.Status}, map[string]string{"status": status, "reason": reason}))
}

func lockPendingApplication(ctx context.Context, q *sqlc.Queries, id int64) (sqlc.OrganizationApplication, error) {
	current, err := q.LockOrganizationApplication(ctx, id)
	if err != nil {
		return current, apperr.NotFound(err, apperr.ApplicationNotFound)
	}
	if current.Status != ApplicationPending {
		return current, apperr.ApplicationStatusInvalid
	}
	return current, nil
}

func createFromApplication(ctx context.Context, q *sqlc.Queries, actor audit_usecase.Actor, app sqlc.OrganizationApplication) (int64, error) {
	org, err := q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{Code: app.Code, Name: app.OrganizationName})
	if repository.IsUniqueViolation(err) {
		return 0, apperr.OrganizationCodeTaken
	}
	if err != nil {
		return 0, err
	}
	in := CreateInput{Actor: actor, Admin: AdminInput{Name: app.AdminName, Email: app.AdminEmail}}
	res, err := createAdmin(ctx, q, org, in, app.PasswordHash)
	if err != nil {
		return 0, err
	}
	return org.ID, recordCreated(ctx, q, actor, res)
}
