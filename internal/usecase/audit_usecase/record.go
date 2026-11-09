package audit_usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/pqtype"

	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Record menyimpan satu baris audit log memakai `q`. Panggil dengan query
// milik transaksi yang sama dengan perubahan datanya.
func Record(ctx context.Context, q *sqlc.Queries, e Entry) error {
	oldValue, err := toJSON(e.OldValue)
	if err != nil {
		return err
	}
	newValue, err := toJSON(e.NewValue)
	if err != nil {
		return err
	}
	return q.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		OrganizationID: e.OrganizationID, UserID: e.ActorID, Action: e.Action,
		Resource: e.Resource, ResourceID: e.ResourceID, OldValue: oldValue, NewValue: newValue,
		IpAddress: optional(e.Client.IP), UserAgent: optional(e.Client.UserAgent),
	})
}

func toJSON(v any) (pqtype.NullRawMessage, error) {
	if v == nil {
		return pqtype.NullRawMessage{}, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return pqtype.NullRawMessage{}, fmt.Errorf("serialisasi audit log: %w", err)
	}
	return pqtype.NullRawMessage{RawMessage: raw, Valid: true}, nil
}

func optional(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
