package partner_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/partner_usecase"
	"ppl-m1-be/internal/usecase/repository/sqlc"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type Handler struct {
	uc *partner_usecase.Usecase
}

type registerRequest struct {
	OrganizationCode string  `json:"organization_code" binding:"required,notblank,max=50"`
	BusinessName     string  `json:"business_name" binding:"required,notblank,max=150"`
	BusinessType     string  `json:"business_type" binding:"required,notblank,max=100"`
	ContactName      string  `json:"contact_name" binding:"required,notblank,max=150"`
	Email            string  `json:"email" binding:"required,email,max=255"`
	Password         string  `json:"password" binding:"required"`
	Phone            string  `json:"phone" binding:"required,notblank,max=30"`
	Address          string  `json:"address" binding:"required,notblank"`
	Description      *string `json:"description"`
}

type statusRequest struct {
	Status string `json:"status" binding:"required,oneof=terverifikasi ditolak ditangguhkan"`
	Reason string `json:"reason"`
}

type partnerResponse struct {
	ID           int64   `json:"id"`
	BusinessName string  `json:"business_name"`
	BusinessType string  `json:"business_type"`
	ContactName  string  `json:"contact_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Address      string  `json:"address"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	StatusReason *string `json:"status_reason"`
	VerifiedAt   *string `json:"verified_at"`
	CreatedAt    string  `json:"created_at"`
}

type publicPartnerResponse struct {
	ID           int64   `json:"id"`
	BusinessName string  `json:"business_name"`
	BusinessType string  `json:"business_type"`
	Description  *string `json:"description"`
}

func toResponse(p sqlc.Partner) partnerResponse {
	return partnerResponse{
		ID: p.ID, BusinessName: p.BusinessName, BusinessType: p.BusinessType, ContactName: p.ContactName,
		Email: p.Email, Phone: p.Phone, Address: p.Address, Description: httpserver.NullStringPtr(p.Description),
		Status: p.Status, StatusReason: httpserver.NullStringPtr(p.StatusReason),
		VerifiedAt: httpserver.NullTimeText(p.VerifiedAt), CreatedAt: httpserver.FormatTime(p.CreatedAt),
	}
}

func toPublic(p sqlc.Partner) publicPartnerResponse {
	return publicPartnerResponse{
		ID: p.ID, BusinessName: p.BusinessName, BusinessType: p.BusinessType, Description: httpserver.NullStringPtr(p.Description),
	}
}

func (r registerRequest) toInput(c *gin.Context) partner_usecase.RegisterInput {
	return partner_usecase.RegisterInput{
		OrganizationCode: r.OrganizationCode, BusinessName: r.BusinessName, BusinessType: r.BusinessType,
		ContactName: r.ContactName, Email: r.Email, Password: r.Password, Phone: r.Phone, Address: r.Address,
		Description: httpserver.OptionalText(r.Description), Client: middleware.Client(c),
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	if !utils.ValidPassword(req.Password) {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"password": httpserver.MsgPasswordRule})
		return
	}
	partner, err := h.uc.Register(c.Request.Context(), req.toInput(c))
	if err != nil {
		respond.Error(c, "daftar partner gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, toResponse(partner))
}

type registrationOrganization struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RegistrationOrganization dipakai form publik untuk menampilkan nama
// organisasi dari tautan. Tidak ada endpoint yang mendaftar semua organisasi.
func (h *Handler) RegistrationOrganization(c *gin.Context) {
	org, err := h.uc.OrganizationForRegistration(c.Request.Context(), c.Param("code"))
	if err != nil {
		respond.Error(c, "ambil organisasi pendaftaran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, registrationOrganization{Code: org.Code, Name: org.Name})
}

// RegistrationLink memberi admin kode organisasinya untuk tautan pendaftaran.
func (h *Handler) RegistrationLink(c *gin.Context) {
	org, err := h.uc.Organization(c.Request.Context(), middleware.Actor(c).OrganizationID)
	if err != nil {
		respond.Error(c, "ambil tautan pendaftaran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, registrationOrganization{Code: org.Code, Name: org.Name})
}

type resubmitRequest struct {
	BusinessName string  `json:"business_name" binding:"required,notblank,max=150"`
	BusinessType string  `json:"business_type" binding:"required,notblank,max=100"`
	ContactName  string  `json:"contact_name" binding:"required,notblank,max=150"`
	Phone        string  `json:"phone" binding:"required,notblank,max=30"`
	Address      string  `json:"address" binding:"required,notblank"`
	Description  *string `json:"description"`
}

// Resubmit dipakai partner yang ditolak untuk mengajukan ulang.
func (h *Handler) Resubmit(c *gin.Context) {
	var req resubmitRequest
	if !httpserver.BindJSON(c, &req) {
		return
	}
	partner, err := h.uc.Resubmit(c.Request.Context(), middleware.Actor(c), partner_usecase.ResubmitInput{
		BusinessName: req.BusinessName, BusinessType: req.BusinessType, ContactName: req.ContactName,
		Phone: req.Phone, Address: req.Address, Description: httpserver.OptionalText(req.Description),
	})
	if err != nil {
		respond.Error(c, "ajukan ulang partner gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(partner))
}

func (h *Handler) Mine(c *gin.Context) {
	partner, err := h.uc.Mine(c.Request.Context(), middleware.Actor(c).UserID)
	if err != nil {
		respond.Error(c, "ambil status partner gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(partner))
}

func (h *Handler) List(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	status, ok := httpserver.QueryEnum(c, "status", partner_usecase.StatusPending, partner_usecase.StatusVerified, partner_usecase.StatusRejected, partner_usecase.StatusSuspended)
	if !ok {
		httpserver.Fail(c, httpserver.ErrInvalidParameter)
		return
	}
	filter := partner_usecase.ListFilter{Status: status, Q: httpserver.QueryString(c, "q")}
	result, err := h.uc.List(c.Request.Context(), middleware.Actor(c).OrganizationID, filter, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "daftar partner gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toResponse), page.Meta(result.Total))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	if !ok {
		return
	}
	partner, err := h.uc.Get(c.Request.Context(), middleware.Actor(c).OrganizationID, id)
	if err != nil {
		respond.Error(c, "ambil partner gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(partner))
}

func (h *Handler) SetStatus(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req statusRequest
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	partner, err := h.uc.SetStatus(c.Request.Context(), middleware.Actor(c), id, req.Status, req.Reason)
	if err != nil {
		respond.Error(c, "ubah status partner gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, toResponse(partner))
}

// ListVerified adalah daftar partner yang tampil di platform.
func (h *Handler) ListVerified(c *gin.Context) {
	partners, err := h.uc.ListVerified(c.Request.Context(), middleware.Actor(c).OrganizationID)
	if err != nil {
		httpserver.Internal(c, "daftar partner terverifikasi gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, httpserver.Map(partners, toPublic))
}
