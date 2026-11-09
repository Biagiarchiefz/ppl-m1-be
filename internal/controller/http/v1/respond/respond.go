// Package respond memetakan galat domain (apperr) ke amplop galat API.
package respond

import (
	"errors"

	"github.com/gin-gonic/gin"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/pkg/httpserver"
)

var simple = []struct {
	err error
	api httpserver.APIError
}{
	{apperr.InvalidCredentials, httpserver.ErrInvalidCredentials},
	{apperr.AccountInactive, httpserver.ErrAccountInactive},
	{apperr.OrganizationInactive, httpserver.ErrOrganizationInactive},
	{apperr.Unauthenticated, httpserver.ErrUnauthenticated},
	{apperr.UserNotFound, httpserver.ErrUserNotFound},
	{apperr.UserAlreadyInactive, httpserver.ErrUserAlreadyInact},
	{apperr.CannotDeactivateSelf, httpserver.ErrCannotDeactSelf},
	{apperr.UserStillActive, httpserver.ErrUserStillActive},
	{apperr.UserAlreadyActive, httpserver.ErrUserAlreadyActive},
	{apperr.UserReferenceInactive, httpserver.ErrUserRefInactive},
	{apperr.RoleNotFound, httpserver.ErrRoleNotFound},
	{apperr.RoleAlreadyInactive, httpserver.ErrRoleAlreadyInactive},
	{apperr.RoleInUse, httpserver.ErrRoleInUse},
	{apperr.RoleSystemLocked, httpserver.ErrRoleSystemLocked},
	{apperr.DivisionNotFound, httpserver.ErrDivisionNotFound},
	{apperr.DivisionAlreadyInactive, httpserver.ErrDivisionAlreadyInact},
	{apperr.DivisionInUse, httpserver.ErrDivisionInUse},
	{apperr.OrganizationNotFound, httpserver.ErrOrgNotFound},
	{apperr.OrganizationAlreadyInactive, httpserver.ErrOrgAlreadyInactive},
	{apperr.OrganizationStillActive, httpserver.ErrOrgStillActive},
	{apperr.RoleAlreadyActive, httpserver.ErrRoleAlreadyActive},
	{apperr.DivisionAlreadyActive, httpserver.ErrDivisionAlreadyActv},
	{apperr.OrganizationAlreadyActive, httpserver.ErrOrgAlreadyActive},
	{apperr.Forbidden, httpserver.ErrForbidden},
	{apperr.ApplicationNotFound, httpserver.ErrApplicationNotFound},
	{apperr.ApplicationPending, httpserver.ErrApplicationPending},
	{apperr.ApplicationStatusInvalid, httpserver.ErrApplicationStatus},
	{apperr.BusinessUnitNotFound, httpserver.ErrBusinessUnitNotFound},
	{apperr.BusinessUnitStatusUnchanged, httpserver.ErrBusinessUnitUnchanged},
	{apperr.PartnerNotFound, httpserver.ErrPartnerNotFound},
	{apperr.PartnerRoleUnavailable, httpserver.ErrPartnerRoleMissing},
	{apperr.PartnerStatusInvalid, httpserver.ErrPartnerStatusInvalid},
	{apperr.PartnerNotVerified, httpserver.ErrPartnerNotVerified},
	{apperr.ProductNotFound, httpserver.ErrProductNotFound},
	{apperr.ProductStatusUnchanged, httpserver.ErrProductUnchanged},
	{apperr.CustomerNotFound, httpserver.ErrCustomerNotFound},
	{apperr.QuotationNotFound, httpserver.ErrQuotationNotFound},
	{apperr.QuotationStatusInvalid, httpserver.ErrQuotationStatusInvalid},
	{apperr.QuotationNotApproved, httpserver.ErrQuotationNotApproved},
	{apperr.InvoiceNotFound, httpserver.ErrInvoiceNotFound},
	{apperr.InvoiceAlreadyExists, httpserver.ErrInvoiceAlreadyExists},
	{apperr.InvoiceAlreadyPaid, httpserver.ErrInvoiceAlreadyPaid},
	{apperr.OrderNotFound, httpserver.ErrOrderNotFound},
	{apperr.OrderCompleted, httpserver.ErrOrderCompleted},
	{apperr.OrderStatusInvalid, httpserver.ErrOrderStatusInvalid},
	{apperr.OutOfStock, httpserver.ErrOutOfStock},
	{apperr.OrderCancelNotAllowed, httpserver.ErrOrderCancelDenied},
	{apperr.OrderAlreadyPaid, httpserver.ErrOrderAlreadyPaid},
	{apperr.InvoiceCancelled, httpserver.ErrInvoiceCancelled},
	{apperr.ServiceRequestNotFound, httpserver.ErrServiceReqNotFound},
	{apperr.ServiceRequestStatusInvalid, httpserver.ErrServiceReqStatus},
	{apperr.CustomerRoleUnavailable, httpserver.ErrCustomerRoleMissing},
	{apperr.CartEmpty, httpserver.ErrCartEmpty},
	{apperr.CartItemNotFound, httpserver.ErrCartItemNotFound},
	{apperr.CartItemUnavailable, httpserver.ErrCartItemUnavailable},
}

// withField: galat yang juga menunjuk satu field di form.
var withField = []struct {
	err   error
	api   httpserver.APIError
	field string
}{
	{apperr.OldPasswordInvalid, httpserver.ErrOldPasswordInvalid, "old_password"},
	{apperr.EmailTaken, httpserver.ErrEmailDuplicate, "email"},
	{apperr.RoleCodeTaken, httpserver.ErrRoleCodeDuplicate, "code"},
	{apperr.DivisionCodeTaken, httpserver.ErrDivisionCodeDup, "code"},
	{apperr.OrganizationCodeTaken, httpserver.ErrOrgCodeDuplicate, "code"},
	{apperr.BusinessUnitCodeTaken, httpserver.ErrBusinessUnitCodeDup, "code"},
	{apperr.CustomerEmailTaken, httpserver.ErrCustomerEmailDuplicate, "email"},
}

// Error mengirim respons galat; galat yang tidak dikenal menjadi 500 dan
// detailnya hanya dicatat di log server.
func Error(c *gin.Context, action string, err error) {
	var fields apperr.FieldErrors
	if errors.As(err, &fields) {
		httpserver.FailFields(c, httpserver.ErrValidation, fields)
		return
	}
	for _, m := range withField {
		if errors.Is(err, m.err) {
			httpserver.FailFields(c, m.api, map[string]string{m.field: m.api.Message})
			return
		}
	}
	for _, m := range simple {
		if errors.Is(err, m.err) {
			httpserver.Fail(c, m.api)
			return
		}
	}
	httpserver.Internal(c, action, err)
}
