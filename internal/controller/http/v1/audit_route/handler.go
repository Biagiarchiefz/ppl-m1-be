package audit_route

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"

	"ppl-m1-be/internal/usecase/audit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

var auditSorts = []string{"-created_at", "created_at"}

type Handler struct {
	uc *audit_usecase.Usecase
}

// auditLogResponse: actor_name, actor_email, dan organization_name adalah
// salinan saat kejadian, tetap terisi walau pengguna/organisasinya sudah
// dihapus permanen (user_id/organization_id lalu menjadi null).
type auditLogResponse struct {
	ID               int64           `json:"id"`
	OrganizationID   *int64          `json:"organization_id"`
	OrganizationName *string         `json:"organization_name"`
	UserID           *int64          `json:"user_id"`
	ActorName        *string         `json:"actor_name"`
	ActorEmail       *string         `json:"actor_email"`
	Action           string          `json:"action"`
	Resource         string          `json:"resource"`
	ResourceID       *int64          `json:"resource_id"`
	OldValue         json.RawMessage `json:"old_value"`
	NewValue         json.RawMessage `json:"new_value"`
	IPAddress        *string         `json:"ip_address"`
	CreatedAt        string          `json:"created_at"`
}

// List menampilkan riwayat audit log organisasi pemanggil (M1-US2). Tidak
// ada endpoint untuk mengubah atau menghapus audit log.
func (h *Handler) List(c *gin.Context) {
	h.list(c, func(f *audit_usecase.ListFilter) bool {
		f.OrganizationID = audit_usecase.ID(middleware.Actor(c).OrganizationID)
		return true
	})
}

// PlatformList menampilkan audit log semua organisasi untuk super admin,
// dengan filter opsional organization_id.
func (h *Handler) PlatformList(c *gin.Context) {
	h.list(c, func(f *audit_usecase.ListFilter) bool {
		orgID, ok := httpserver.QueryInt64(c, "organization_id")
		f.AllOrganizations, f.OrganizationID = true, orgID
		return ok
	})
}

// list menjalankan daftar audit log. scope menetapkan organisasi yang boleh
// dilihat dan mengembalikan false bila parameternya tidak valid.
func (h *Handler) list(c *gin.Context, scope func(*audit_usecase.ListFilter) bool) {
	page, okPage := httpserver.ParsePage(c)
	filter, okFilter := parseFilter(c)
	if okScope := scope(&filter); !okPage || !okFilter || !okScope {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.List(c.Request.Context(), filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar audit log gagal", err)
		return
	}
	httpserver.SuccessList(c, toResponses(result.Items), page.Meta(result.Total))
}

// parseFilter: `from` dan `to` (YYYY-MM-DD, WIB) keduanya ikut dihitung,
// jadi batas atas adalah awal hari setelah `to`.
func parseFilter(c *gin.Context) (audit_usecase.ListFilter, bool) {
	from, okFrom := httpserver.QueryDate(c, "from")
	to, okTo := httpserver.QueryDate(c, "to")
	userID, okUser := httpserver.QueryInt64(c, "user_id")
	sort, okSort := httpserver.QuerySort(c, auditSorts, "-created_at")
	if from.Valid && to.Valid && to.Time.Before(from.Time) {
		return audit_usecase.ListFilter{}, false
	}
	if to.Valid {
		to.Time = to.Time.Add(24 * time.Hour)
	}
	return audit_usecase.ListFilter{
		From: from, To: to, UserID: userID, Sort: sort,
		Action:   httpserver.QueryString(c, "action"),
		Resource: httpserver.QueryString(c, "resource"),
	}, okFrom && okTo && okUser && okSort
}

func toResponses(rows []sqlc.ListAuditLogsRow) []auditLogResponse {
	out := make([]auditLogResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, auditLogResponse{
			ID: r.ID, OrganizationID: httpserver.NullInt64Ptr(r.OrganizationID),
			OrganizationName: httpserver.NullStringPtr(r.OrganizationName), UserID: httpserver.NullInt64Ptr(r.UserID),
			ActorName: httpserver.NullStringPtr(r.ActorName), ActorEmail: httpserver.NullStringPtr(r.ActorEmail),
			Action: r.Action, Resource: r.Resource, ResourceID: httpserver.NullInt64Ptr(r.ResourceID),
			OldValue: rawOrNull(r.OldValue), NewValue: rawOrNull(r.NewValue),
			IPAddress: httpserver.NullStringPtr(r.IpAddress), CreatedAt: httpserver.FormatTime(r.CreatedAt),
		})
	}
	return out
}

func rawOrNull(v pqtype.NullRawMessage) json.RawMessage {
	if !v.Valid {
		return json.RawMessage("null")
	}
	return v.RawMessage
}
