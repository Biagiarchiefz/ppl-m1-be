package organization_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/organization_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type applicationRequest struct {
	Code             string  `json:"code" binding:"required"`
	OrganizationName string  `json:"organization_name" binding:"required,notblank,max=150"`
	AdminName        string  `json:"admin_name" binding:"required,notblank,max=150"`
	AdminEmail       string  `json:"admin_email" binding:"required,email,max=255"`
	Password         string  `json:"password" binding:"required"`
	Phone            string  `json:"phone" binding:"required,notblank,max=30"`
	Description      *string `json:"description"`
}

type lookupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type reviewRequest struct {
	Status string `json:"status" binding:"required,oneof=disetujui ditolak"`
	Reason string `json:"reason"`
}

type applicationResponse struct {
	ID               int64   `json:"id"`
	Code             string  `json:"code"`
	OrganizationName string  `json:"organization_name"`
	AdminName        string  `json:"admin_name"`
	AdminEmail       string  `json:"admin_email"`
	Phone            string  `json:"phone"`
	Description      *string `json:"description"`
	Status           string  `json:"status"`
	StatusReason     *string `json:"status_reason"`
	ReviewedAt       *string `json:"reviewed_at"`
	CreatedAt        string  `json:"created_at"`
}

func toApplicationResponse(a sqlc.OrganizationApplication) applicationResponse {
	return applicationResponse{
		ID: a.ID, Code: a.Code, OrganizationName: a.OrganizationName, AdminName: a.AdminName, AdminEmail: a.AdminEmail, Phone: a.Phone,
		Description: httpserver.NullStringPtr(a.Description), Status: a.Status, StatusReason: httpserver.NullStringPtr(a.StatusReason),
		ReviewedAt: httpserver.NullTimeText(a.ReviewedAt), CreatedAt: httpserver.FormatTime(a.CreatedAt),
	}
}

func applicationFieldErrors(req applicationRequest) map[string]string {
	fields := map[string]string{}
	if !utils.ValidCode(req.Code) {
		fields["code"] = httpserver.MsgCodeFormat
	}
	if !utils.ValidPassword(req.Password) {
		fields["password"] = httpserver.MsgPasswordRule
	}
	return fields
}

func (h *Handler) Apply(c *gin.Context) {
	var req applicationRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	if fields := applicationFieldErrors(req); len(fields) > 0 {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return
	}
	application, err := h.uc.Apply(c.Request.Context(), organization_usecase.ApplicationInput{
		Code: req.Code, OrganizationName: req.OrganizationName, AdminName: req.AdminName, AdminEmail: req.AdminEmail,
		Password: req.Password, Phone: req.Phone, Description: httpserver.OptionalText(req.Description), Client: middleware.Client(c),
	})
	if err != nil {
		respond.Error(c, "ajukan organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toApplicationResponse(application))
}

// LookupApplication memakai POST agar kata sandi tidak muncul di alamat.
func (h *Handler) LookupApplication(c *gin.Context) {
	var req lookupRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	application, err := h.uc.ApplicationStatus(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respond.Error(c, "cek pengajuan organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toApplicationResponse(application))
}

func (h *Handler) ListApplications(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	status, ok := httpserver.QueryEnum(c, "status", organization_usecase.ApplicationPending, organization_usecase.ApplicationApproved, organization_usecase.ApplicationRejected)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	result, err := h.uc.ListApplications(c.Request.Context(), status, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar pengajuan organisasi gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toApplicationResponse), page.Meta(result.Total))
}

func (h *Handler) ReviewApplication(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req reviewRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	application, err := h.uc.ReviewApplication(c.Request.Context(), middleware.Actor(c), id, req.Status, req.Reason)
	if err != nil {
		respond.Error(c, "tinjau pengajuan organisasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toApplicationResponse(application))
}
