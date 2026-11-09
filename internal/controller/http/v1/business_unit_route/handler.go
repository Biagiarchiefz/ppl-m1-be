package business_unit_route

import (
	"database/sql"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/business_unit_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *business_unit_usecase.Usecase
}

type unitFields struct {
	DivisionID  *int64  `json:"division_id"`
	Name        string  `json:"name" binding:"required,notblank,max=150"`
	Category    string  `json:"category" binding:"required,notblank,max=100"`
	Description *string `json:"description"`
	Address     *string `json:"address"`
	Phone       *string `json:"phone" binding:"omitempty,max=30"`
	Email       *string `json:"email" binding:"omitempty,email,max=255"`
}

type createRequest struct {
	unitFields
	Code string `json:"code" binding:"required"`
}

type metricRequest struct {
	BusinessUnitID int64   `json:"business_unit_id" binding:"required,gt=0"`
	SourceModule   string  `json:"source_module" binding:"required,notblank,max=20"`
	MetricKey      string  `json:"metric_key" binding:"required,notblank,max=50"`
	Label          string  `json:"label" binding:"required,notblank,max=100"`
	Unit           string  `json:"unit" binding:"required,notblank,max=20"`
	Value          *int64  `json:"value" binding:"required"`
	PeriodDate     *string `json:"period_date"`
}

type unitResponse struct {
	ID           int64   `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	DivisionID   *int64  `json:"division_id"`
	DivisionName *string `json:"division_name"`
	Description  *string `json:"description"`
	Address      *string `json:"address"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// ToResponse dipakai juga oleh katalog M5.
func ToResponse(b sqlc.BusinessUnitDetail) unitResponse {
	return unitResponse{
		ID: b.ID, Code: b.Code, Name: b.Name, Category: b.Category,
		DivisionID: httpserver.NullInt64Ptr(b.DivisionID), DivisionName: httpserver.NullStringPtr(b.DivisionName),
		Description: httpserver.NullStringPtr(b.Description), Address: httpserver.NullStringPtr(b.Address),
		Phone: httpserver.NullStringPtr(b.Phone), Email: httpserver.NullStringPtr(b.Email), IsActive: b.IsActive,
		CreatedAt: httpserver.FormatTime(b.CreatedAt), UpdatedAt: httpserver.FormatTime(b.UpdatedAt),
	}
}

func (f unitFields) toInput(code string) business_unit_usecase.Input {
	return business_unit_usecase.Input{
		DivisionID: httpserver.OptionalID(f.DivisionID), Code: code, Name: f.Name, Category: f.Category,
		Description: httpserver.OptionalText(f.Description), Address: httpserver.OptionalText(f.Address),
		Phone: httpserver.OptionalText(f.Phone), Email: httpserver.OptionalText(f.Email),
	}
}

// listFilter memaksa hanya unit aktif untuk pemanggil tanpa izin kelola,
// sehingga unit nonaktif tidak muncul di pilihan modul lain.
func listFilter(c *gin.Context) (business_unit_usecase.ListFilter, bool) {
	isActive, ok := httpserver.QueryBool(c, "is_active")
	if !slices.Contains(middleware.GetClaims(c).Permissions, utils.PermBusinessUnitsManage) {
		isActive = sql.NullBool{Bool: true, Valid: true}
	}
	return business_unit_usecase.ListFilter{IsActive: isActive, Q: httpserver.QueryString(c, "q")}, ok
}

func (h *Handler) List(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	filter, ok := listFilter(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.List(c.Request.Context(), middleware.Actor(c).OrganizationID, filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar unit usaha gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, ToResponse), page.Meta(result.Total))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	if !ok {
		return
	}
	unit, err := h.uc.Get(c.Request.Context(), middleware.Actor(c).OrganizationID, id)
	if err != nil {
		respond.Error(c, "ambil unit usaha gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, ToResponse(unit))
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	if !utils.ValidCode(req.Code) {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"code": httpserver.MsgCodeFormat})
		return
	}
	unit, err := h.uc.Create(c.Request.Context(), middleware.Actor(c), req.toInput(req.Code))
	if err != nil {
		respond.Error(c, "tambah unit usaha gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, ToResponse(unit))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req unitFields
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	unit, err := h.uc.Update(c.Request.Context(), middleware.Actor(c), id, req.toInput(""))
	if err != nil {
		respond.Error(c, "ubah unit usaha gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, ToResponse(unit))
}

func (h *Handler) SetStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetStatus(ctx, middleware.Actor(c), id, active); err != nil {
		respond.Error(c, "ubah status unit usaha gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}

func periodDate(raw *string) (sql.NullTime, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return sql.NullTime{}, true
	}
	t, err := time.ParseInLocation("2006-01-02", *raw, httpserver.WIB)
	return sql.NullTime{Time: t, Valid: err == nil}, err == nil
}

func (h *Handler) SubmitMetric(c *gin.Context) {
	var req metricRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	date, ok := periodDate(req.PeriodDate)
	if !ok {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"period_date": "Format tanggal harus YYYY-MM-DD"})
		return
	}
	err := h.uc.Submit(c.Request.Context(), business_unit_usecase.Metric{
		OrganizationID: middleware.Actor(c).OrganizationID, BusinessUnitID: sql.NullInt64{Int64: req.BusinessUnitID, Valid: true}, SourceModule: req.SourceModule,
		Key: req.MetricKey, Label: req.Label, Unit: req.Unit, Value: *req.Value, PeriodDate: date,
	})
	if err != nil {
		respond.Error(c, "kirim data dashboard gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, gin.H{"received": true})
}

type reportItem struct {
	BusinessUnitID   *int64 `json:"business_unit_id"`
	BusinessUnitName string `json:"business_unit_name"`
	SourceModule     string `json:"source_module"`
	MetricKey        string `json:"metric_key"`
	Label            string `json:"label"`
	Unit             string `json:"unit"`
	Value            int64  `json:"value"`
}

func toReportItem(r sqlc.ReportDashboardMetricsRow) reportItem {
	return reportItem{
		BusinessUnitID: httpserver.NullInt64Ptr(r.BusinessUnitID), BusinessUnitName: r.BusinessUnitName, SourceModule: r.SourceModule,
		MetricKey: r.MetricKey, Label: r.Label, Unit: r.Unit, Value: r.Total,
	}
}

// reportUnit mengunci pengguna yang terikat unit usaha ke unitnya sendiri.
// Administrator tetap bebas memilih unit lewat filter.
func reportUnit(c *gin.Context) (sql.NullInt64, bool) {
	claims := middleware.GetClaims(c)
	if claims.UnitUsahaID != nil && claims.Role != utils.SystemRoleCode {
		return sql.NullInt64{Int64: *claims.UnitUsahaID, Valid: true}, true
	}
	return httpserver.QueryInt64(c, "business_unit_id")
}

func reportFilter(c *gin.Context) (business_unit_usecase.ReportFilter, bool) {
	unit, okUnit := reportUnit(c)
	from, okFrom := httpserver.QueryDate(c, "from")
	to, okTo := httpserver.QueryDate(c, "to")
	valid := okUnit && okFrom && okTo && (!from.Valid || !to.Valid || !from.Time.After(to.Time))
	return business_unit_usecase.ReportFilter{BusinessUnitID: unit, From: from, To: to}, valid
}

// Report mengembalikan `items: []` bila belum ada data; bukan galat.
func (h *Handler) Report(c *gin.Context) {
	filter, ok := reportFilter(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	rows, err := h.uc.Report(c.Request.Context(), middleware.Actor(c).OrganizationID, filter)
	if err != nil {
		httpserver.Internal(c, "ambil dashboard gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"items": httpserver.Map(rows, toReportItem)})
}
