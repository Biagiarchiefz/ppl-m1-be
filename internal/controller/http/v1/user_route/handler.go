package user_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/user_usecase"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

var userSorts = []string{"created_at", "-created_at", "name", "-name", "email", "-email"}

type Handler struct {
	uc *user_usecase.Usecase
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

	user, err := h.uc.Create(c.Request.Context(), user_usecase.CreateInput{
		Actor: middleware.Actor(c), Name: req.Name, Email: req.Email, Password: req.Password, Role: req.Role,
		Division: toNullString(req.Division), UnitUsahaID: toNullInt64(req.UnitUsahaID),
	})
	if err != nil {
		respond.Error(c, "tambah akun gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toUserResponse(user))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	user, err := h.uc.Get(c.Request.Context(), middleware.Actor(c), id)
	if err != nil {
		respond.Error(c, "ambil akun gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toUserResponse(user))
}

func (h *Handler) Update(c *gin.Context) {
	id, req, ok := bindUpdate(c)
	if !ok {
		return
	}
	in := user_usecase.UpdateInput{Actor: middleware.Actor(c), ID: id}
	toUpdateFields(req, &in)

	user, err := h.uc.Update(c.Request.Context(), in)
	if err != nil {
		respond.Error(c, "ubah akun gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toUserResponse(user))
}

func bindUpdate(c *gin.Context) (int64, updateRequest, bool) {
	var req updateRequest
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return 0, req, false
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.BindError(c, err)
		return 0, req, false
	}
	if fields := updateFieldErrors(req); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return 0, req, false
	}
	return id, req, true
}

func (h *Handler) Deactivate(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Deactivate(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "nonaktifkan akun gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, deactivateResponse{ID: id, IsActive: false})
}

// SetStatus mengaktifkan atau menonaktifkan akun, tanpa menghapus riwayatnya.
func (h *Handler) SetStatus(c *gin.Context) {
	ctx, id, active, ok := httpserver.StatusUpdate(c)
	if !ok {
		return
	}
	if err := h.uc.SetStatus(ctx, middleware.Actor(c), id, active); err != nil {
		respond.Error(c, "ubah status akun gagal", err)
		return
	}
	httpserver.StatusUpdated(c, id, active)
}

func (h *Handler) List(c *gin.Context) {
	page, okPage := httpserver.ParsePage(c)
	filter, okFilter := parseUserFilter(c)
	if !okPage || !okFilter {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.List(c.Request.Context(), middleware.Actor(c), filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar akun gagal", err)
		return
	}
	httpserver.SuccessList(c, toUserList(result.Items), page.Meta(result.Total))
}

func parseUserFilter(c *gin.Context) (user_usecase.ListFilter, bool) {
	unitUsahaID, okUnit := httpserver.QueryInt64(c, "unit_usaha_id")
	isActive, okActive := httpserver.QueryBool(c, "is_active")
	sort, okSort := httpserver.QuerySort(c, userSorts, "-created_at")
	return user_usecase.ListFilter{
		Q:           httpserver.QueryString(c, "q"),
		Role:        httpserver.QueryString(c, "role"),
		Division:    httpserver.QueryString(c, "division"),
		UnitUsahaID: unitUsahaID,
		IsActive:    isActive,
		Sort:        sort,
	}, okUnit && okActive && okSort
}

// Delete menghapus permanen akun yang sudah nonaktif.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := httpserver.PathID(c)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	if err := h.uc.Delete(c.Request.Context(), middleware.Actor(c), id); err != nil {
		respond.Error(c, "hapus permanen akun gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "deleted": true})
}
