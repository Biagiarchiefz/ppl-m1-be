package audit_usecase

import (
	"database/sql"

	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Nama aksi dan resource audit log.
const (
	ActionUserCreated         = "user.created"
	ActionUserUpdated         = "user.updated"
	ActionUserActivated       = "user.activated"
	ActionUserDeactivated     = "user.deactivated"
	ActionUserDeleted         = "user.deleted"
	ActionUserPasswordChanged = "user.password_changed"
	ActionLoginSucceeded      = "auth.login_succeeded"
	ActionLoginFailed         = "auth.login_failed"
	ActionRoleCreated         = "role.created"
	ActionRoleUpdated         = "role.updated"
	ActionRoleActivated       = "role.activated"
	ActionRoleDeactivated     = "role.deactivated"
	ActionDivisionCreated     = "division.created"
	ActionDivisionUpdated     = "division.updated"
	ActionDivisionActivated   = "division.activated"
	ActionDivisionDeactivated = "division.deactivated"
	ActionOrgCreated          = "organization.created"
	ActionOrgUpdated          = "organization.updated"
	ActionOrgActivated        = "organization.activated"
	ActionOrgDeactivated      = "organization.deactivated"
	ActionOrgDeleted          = "organization.deleted"

	ActionBusinessUnitCreated = "business_unit.created"
	ActionBusinessUnitUpdated = "business_unit.updated"
	ActionBusinessUnitStatus  = "business_unit.status_changed"
	ActionPartnerRegistered   = "partner.registered"
	ActionPartnerStatus       = "partner.status_changed"
	ActionProductCreated      = "product.created"
	ActionProductUpdated      = "product.updated"
	ActionProductStatus       = "product.status_changed"
	ActionCustomerCreated     = "customer.created"
	ActionCustomerUpdated     = "customer.updated"
	ActionQuotationCreated    = "quotation.created"
	ActionQuotationStatus     = "quotation.status_changed"
	ActionInvoiceCreated      = "invoice.created"
	ActionInvoicePaid         = "invoice.paid"
	ActionPaymentFailed       = "payment.failed"
	ActionOrderCreated        = "order.created"
	ActionOrderStatus         = "order.status_changed"
	ActionServiceRequestSent  = "service_request.created"

	ActionApplicationSubmitted = "organization_application.submitted"
	ActionApplicationReviewed  = "organization_application.reviewed"
	ResourceApplications       = "organization_applications"

	ActionPartnerResubmitted   = "partner.resubmitted"
	ActionCustomerRegistered   = "customer.registered"
	ActionOrderCancelled       = "order.cancelled"
	ActionServiceRequestStatus = "service_request.status_changed"

	ResourceUsers           = "users"
	ResourceAuth            = "auth"
	ResourceRoles           = "roles"
	ResourceDivisions       = "divisions"
	ResourceOrganizations   = "organizations"
	ResourceBusinessUnits   = "business_units"
	ResourcePartners        = "partners"
	ResourceProducts        = "products"
	ResourceCustomers       = "customers"
	ResourceQuotations      = "quotations"
	ResourceInvoices        = "invoices"
	ResourceOrders          = "orders"
	ResourceServiceRequests = "service_requests"
)

// Client adalah asal permintaan yang ikut dicatat di audit log.
type Client struct {
	IP        string
	UserAgent string
}

// Actor adalah pengguna yang menjalankan aksi. OrganizationID = 0 untuk
// super admin platform.
type Actor struct {
	UserID         int64
	OrganizationID int64
	Client         Client
}

// Entry adalah satu baris audit log. OldValue/NewValue tidak boleh memuat
// password, password_hash, atau token.
type Entry struct {
	OrganizationID sql.NullInt64
	ActorID        sql.NullInt64
	Action         string
	Resource       string
	ResourceID     sql.NullInt64
	OldValue       any
	NewValue       any
	Client         Client
}

// Change membuat Entry untuk aksi `actor` pada data milik organisasi `orgID`.
func (a Actor) Change(orgID int64, action, resource string, resourceID int64, oldValue, newValue any) Entry {
	return Entry{
		OrganizationID: ID(orgID), ActorID: ID(a.UserID),
		Action: action, Resource: resource, ResourceID: ID(resourceID),
		OldValue: oldValue, NewValue: newValue, Client: a.Client,
	}
}

// ListFilter: AllOrganizations = true hanya untuk super admin; OrganizationID
// lalu menjadi filter opsional. Selain itu OrganizationID wajib diisi.
type ListFilter struct {
	AllOrganizations bool
	OrganizationID   sql.NullInt64
	From             sql.NullTime
	To               sql.NullTime
	UserID           sql.NullInt64
	Action           sql.NullString
	Resource         sql.NullString
	Sort             string
}

type ListResult struct {
	Items []sqlc.ListAuditLogsRow
	Total int64
}

// ID mengubah id menjadi nilai nullable; 0 dianggap kosong.
func ID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id != 0}
}
