package storefront_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// StoreProfile adalah isi toko publik satu organisasi.
type StoreProfile struct {
	Organization sqlc.Organization
	Units        []sqlc.BusinessUnitDetail
	Partners     []sqlc.Partner
	Categories   []string
}

// Store dibuka tanpa login. Organisasi nonaktif dianggap tidak ada.
func (u *Usecase) Store(ctx context.Context, code string) (StoreProfile, error) {
	org, err := u.OrganizationForRegistration(ctx, code)
	if err != nil {
		return StoreProfile{}, err
	}
	profile := StoreProfile{Organization: org}
	if profile.Units, err = u.store.ListCatalogBusinessUnits(ctx, org.ID); err != nil {
		return StoreProfile{}, err
	}
	if profile.Partners, err = u.store.ListVerifiedPartners(ctx, org.ID); err != nil {
		return StoreProfile{}, err
	}
	profile.Categories, err = u.store.ListCatalogCategories(ctx, org.ID)
	return profile, err
}

// StoreProducts memakai aturan tampil yang sama dengan katalog setelah login.
func (u *Usecase) StoreProducts(ctx context.Context, code string, f CatalogFilter, limit, offset int32) (ProductPage, error) {
	org, err := u.OrganizationForRegistration(ctx, code)
	if err != nil {
		return ProductPage{}, err
	}
	return u.Products(ctx, org.ID, f, limit, offset)
}
