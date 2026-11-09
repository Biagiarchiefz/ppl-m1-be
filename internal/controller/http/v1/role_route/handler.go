package role_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/role_usecase"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *role_usecase.Usecase
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
		httpserver.Internal(c, "daftar peran gagal", err)
		return
	}
	httpserver.SuccessList(c, toRoleList(result.Items), page.Meta(result.Total))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	view, err := h.uc.Get(c.Request.Context(), middleware.Actor(c), id)
	if err != nil {
		respond.Error(c, "ambil peran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toRoleResponse(view))
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
	view, err := h.uc.Create(c.Request.Context(), role_usecase.CreateInput{
		Actor: middleware.Actor(c), Code: req.Code, Name: req.Name,
		Description: toDescription(req.Description), Permissions: req.Permissions,
	})
	if err != nil {
		respond.Error(c, "tambah peran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toRoleResponse(view))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	var req updateRequest
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return
	}
	if fields := updateFieldErrors(req); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return
	}
	in := role_usecase.UpdateInput{Actor: middleware.Actor(c), ID: id}
	toUpdateInput(req, &in)
	h.respondUpdate(c, in)
}

func (h *Handler) respondUpdate(c *gin.Context, in role_usecase.UpdateInput) {
	view, err := h.uc.Update(c.Request.Context(), in)
	if err != nil {
		respond.Error(c, "ubah peran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toRoleResponse(view))
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Deactivate(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "nonaktifkan peran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "is_active": false})
}

// SetStatus mengaktifkan atau menonaktifkan peran.
func (h *Handler) SetStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetStatus(ctx, middleware.Actor(c), id, active); err != nil {
		respond.Error(c, "ubah status peran gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}

// Permissions menampilkan seluruh katalog izin (tanpa pagination karena
// jumlahnya kecil dan dibutuhkan utuh untuk form peran).
func (h *Handler) Permissions(c *gin.Context) {
	items, err := h.uc.Permissions(c.Request.Context())
	if err != nil {
		httpserver.Internal(c, "daftar izin gagal", err)
		return
	}
	total := int64(len(items))
	httpserver.SuccessList(c, toPermissionList(items), httpserver.Meta{
		Page: 1, PerPage: len(items), Total: total, TotalPages: min(total, 1),
	})
}
