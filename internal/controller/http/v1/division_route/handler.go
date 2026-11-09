package division_route

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/division_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *division_usecase.Usecase
}

type createRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type renameRequest struct {
	Name string `json:"name" binding:"required"`
}

type divisionResponse struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func toResponse(d sqlc.Division) divisionResponse {
	return divisionResponse{ID: d.ID, Code: d.Code, Name: d.Name, IsActive: d.IsActive, CreatedAt: httpserver.FormatTime(d.CreatedAt)}
}

func nameError(name string, fields map[string]string) {
	if n := utf8.RuneCountInString(strings.TrimSpace(name)); n == 0 || n > 100 {
		fields["name"] = "Wajib diisi, maksimal 100 karakter"
	}
}

func (h *Handler) List(c *gin.Context) {
	page, okPage := httpserver.ParsePage(c)
	isActive, okActive := httpserver.QueryBool(c, "is_active")
	if !okPage || !okActive {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.List(c.Request.Context(), middleware.Actor(c), isActive, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar divisi gagal", err)
		return
	}
	items := make([]divisionResponse, 0, len(result.Items))
	for _, d := range result.Items {
		items = append(items, toResponse(d))
	}
	httpserver.SuccessList(c, items, page.Meta(result.Total))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	d, err := h.uc.Get(c.Request.Context(), middleware.Actor(c), id)
	if err != nil {
		respond.Error(c, "ambil divisi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(d))
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	fields := map[string]string{}
	if !utils.ValidCode(req.Code) {
		fields["code"] = httpserver.MsgCodeFormat
	}
	if nameError(req.Name, fields); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return
	}
	d, err := h.uc.Create(c.Request.Context(), middleware.Actor(c), req.Code, req.Name)
	if err != nil {
		respond.Error(c, "tambah divisi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toResponse(d))
}

func bindRename(c *gin.Context) (int64, string, bool) {
	id, ok := httpserver.PathID(c)
	var req renameRequest
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return 0, "", false
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return 0, "", false
	}
	fields := map[string]string{}
	if nameError(req.Name, fields); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return 0, "", false
	}
	return id, req.Name, true
}

func (h *Handler) Rename(c *gin.Context) {
	id, name, ok := bindRename(c)
	if !ok {
		return
	}
	d, err := h.uc.Rename(c.Request.Context(), middleware.Actor(c), id, name)
	if err != nil {
		respond.Error(c, "ubah divisi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(d))
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Deactivate(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "nonaktifkan divisi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "is_active": false})
}

// SetStatus mengaktifkan atau menonaktifkan divisi.
func (h *Handler) SetStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetStatus(ctx, middleware.Actor(c), id, active); err != nil {
		respond.Error(c, "ubah status divisi gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}
