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
