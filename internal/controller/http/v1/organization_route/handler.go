package organization_route

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/organization_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *organization_usecase.Usecase
}

type adminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type createRequest struct {
	Code  string       `json:"code" binding:"required"`
	Name  string       `json:"name" binding:"required"`
	Admin adminRequest `json:"admin" binding:"required"`
}

type renameRequest struct {
	Name string `json:"name" binding:"required"`
}

type organizationResponse struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type createResponse struct {
	organizationResponse
	Admin struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	} `json:"admin"`
}

func toResponse(o sqlc.Organization) organizationResponse {
	return organizationResponse{ID: o.ID, Code: o.Code, Name: o.Name, IsActive: o.IsActive, CreatedAt: httpserver.FormatTime(o.CreatedAt)}
}

func createFieldErrors(req createRequest) map[string]string {
	fields := map[string]string{}
	if !utils.ValidCode(req.Code) {
		fields["code"] = httpserver.MsgCodeFormat
	}
	nameError("name", req.Name, fields)
	nameError("admin.name", req.Admin.Name, fields)
	if !utils.ValidPassword(req.Admin.Password) {
		fields["admin.password"] = httpserver.MsgPasswordRule
	}
	return fields
}

func nameError(field, name string, fields map[string]string) {
	if n := utf8.RuneCountInString(strings.TrimSpace(name)); n == 0 || n > 150 {
		fields[field] = "Wajib diisi, maksimal 150 karakter"
	}
}

func (h *Handler) List(c *gin.Context) {
	page, okPage := httpserver.ParsePage(c)
	isActive, okActive := httpserver.QueryBool(c, "is_active")
	if !okPage || !okActive {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	filter := organization_usecase.ListFilter{Q: httpserver.QueryString(c, "q"), IsActive: isActive}
	result, err := h.uc.List(c.Request.Context(), filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar organisasi gagal", err)
		return
	}
	items := make([]organizationResponse, 0, len(result.Items))
	for _, o := range result.Items {
		items = append(items, toResponse(o))
	}
	httpserver.SuccessList(c, items, page.Meta(result.Total))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	org, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		respond.Error(c, "ambil organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(org))
}

func (h *Handler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	if fields := createFieldErrors(req); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), organization_usecase.CreateInput{
		Actor: middleware.Actor(c), Code: req.Code, Name: req.Name,
		Admin: organization_usecase.AdminInput{Name: req.Admin.Name, Email: req.Admin.Email, Password: req.Admin.Password},
	})
	if err != nil {
		respond.Error(c, "tambah organisasi gagal", err)
		return
	}
	out := createResponse{organizationResponse: toResponse(res.Organization)}
	out.Admin.ID, out.Admin.Email = res.AdminID, res.AdminEmail
	httpserver.Success(c, http.StatusCreated, out)
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
	if nameError("name", req.Name, fields); len(fields) > 0 {
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
	org, err := h.uc.Rename(c.Request.Context(), middleware.Actor(c), id, name)
	if err != nil {
		respond.Error(c, "ubah organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(org))
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Deactivate(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "nonaktifkan organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "is_active": false})
}

// SetStatus mengaktifkan atau menonaktifkan organisasi.
func (h *Handler) SetStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetStatus(ctx, middleware.Actor(c), id, active); err != nil {
		respond.Error(c, "ubah status organisasi gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}

// Delete menghapus permanen organisasi yang sudah nonaktif beserta
// pengguna, peran, dan divisinya.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Delete(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "hapus permanen organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "deleted": true})
}
