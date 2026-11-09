package storefront_route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/controller/http/v1/respond"
	"ppl-m1-be/internal/usecase/storefront_usecase"
	"ppl-m1-be/internal/usecase/utils"
	"ppl-m1-be/middleware"
	"ppl-m1-be/pkg/httpserver"
)

type customerRegistrationBody struct {
	OrganizationCode string `json:"organization_code" binding:"required,notblank,max=50"`
	Name             string `json:"name" binding:"required,notblank,max=150"`
	Email            string `json:"email" binding:"required,email,max=255"`
	Password         string `json:"password" binding:"required"`
	Phone            string `json:"phone" binding:"required,notblank,max=30"`
}

type serviceRequestStatusBody struct {
	Status string `json:"status" binding:"required,oneof=diproses selesai"`
}

// RegistrationOrganization menampilkan nama organisasi dari tautan pendaftaran.
func (h *Handler) RegistrationOrganization(c *gin.Context) {
	org, err := h.uc.OrganizationForRegistration(c.Request.Context(), c.Param("code"))
	if err != nil {
		respond.Error(c, "ambil organisasi pendaftaran gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"code": org.Code, "name": org.Name})
}

func (h *Handler) RegisterCustomer(c *gin.Context) {
	var req customerRegistrationBody
	if !httpserver.BindJSON(c, &req) {
		return
	}
	if !utils.ValidPassword(req.Password) {
		httpserver.FailFields(c, httpserver.ErrValidation, map[string]string{"password": httpserver.MsgPasswordRule})
		return
	}
	err := h.uc.RegisterCustomer(c.Request.Context(), storefront_usecase.CustomerRegistration{
		OrganizationCode: req.OrganizationCode, Name: req.Name, Email: req.Email, Password: req.Password, Phone: req.Phone, Client: middleware.Client(c),
	})
	if err != nil {
		respond.Error(c, "daftar customer gagal", err)
		return
	}
	httpserver.Success(c, http.StatusCreated, gin.H{"registered": true})
}

func (h *Handler) SetServiceRequestStatus(c *gin.Context) {
	id, ok := httpserver.RequireID(c)
	var req serviceRequestStatusBody
	if !ok || !httpserver.BindJSON(c, &req) {
		return
	}
	unit, _ := serviceRequestUnit(c)
	if err := h.uc.SetServiceRequestStatus(c.Request.Context(), middleware.Actor(c), unit, id, req.Status); err != nil {
		respond.Error(c, "ubah status permintaan layanan gagal", err)
		return
	}
	httpserver.Success(c, http.StatusOK, gin.H{"id": id, "status": req.Status})
}

func (h *Handler) MyServiceRequests(c *gin.Context) {
	page, ok := httpserver.RequirePage(c)
	if !ok {
		return
	}
	result, err := h.uc.MyServiceRequests(c.Request.Context(), middleware.Actor(c).UserID, page.Limit(), page.Offset())
	if err != nil {
		httpserver.Internal(c, "riwayat permintaan layanan gagal", err)
		return
	}
	httpserver.SuccessList(c, httpserver.Map(result.Items, toServiceRequest), page.Meta(result.Total))
}
