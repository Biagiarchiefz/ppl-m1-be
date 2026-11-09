package user_usecase

import (
	"context"
	"database/sql"
	"errors"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

const (
	msgRoleUnknown     = "Peran tidak terdaftar"
	msgDivisionUnknown = "Divisi tidak terdaftar"
)

// refs adalah id peran/divisi hasil pencarian kode di organisasi pemanggil.
type refs struct {
	RoleID     int64
	DivisionID sql.NullInt64
}

// resolveRefs mengubah kode peran/divisi menjadi id. Kode yang tidak ada
// (atau nonaktif) di organisasi tersebut menjadi galat 422 per field.
// Peran kosong dan divisi null dilewati.
func resolveRefs(ctx context.Context, q *sqlc.Queries, orgID int64, role string, division sql.NullString) (refs, error) {
	var out refs
	fields := apperr.FieldErrors{}
	if role != "" {
		id, err := q.GetActiveRoleIDByCode(ctx, sqlc.GetActiveRoleIDByCodeParams{OrganizationID: orgID, Code: role})
		if err := lookupError(err, fields, "role", msgRoleUnknown); err != nil {
			return refs{}, err
		}
		out.RoleID = id
	}
	if division.Valid {
		id, err := q.GetActiveDivisionIDByCode(ctx, sqlc.GetActiveDivisionIDByCodeParams{OrganizationID: orgID, Code: division.String})
		if err := lookupError(err, fields, "division", msgDivisionUnknown); err != nil {
			return refs{}, err
		}
		out.DivisionID = sql.NullInt64{Int64: id, Valid: id != 0}
	}
	if len(fields) > 0 {
		return refs{}, fields
	}
	return out, nil
}

func lookupError(err error, fields apperr.FieldErrors, field, message string) error {
	if errors.Is(err, sql.ErrNoRows) {
		fields[field] = message
		return nil
	}
	return err
}
