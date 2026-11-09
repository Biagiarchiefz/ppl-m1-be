package auth_usecase

import (
	"context"

	"ppl-m1-be/internal/usecase/apperr"
)

// Me mengambil profil terbaru dari database (K-01), sehingga akun atau
// organisasi yang dinonaktifkan setelah token terbit tetap ditolak di sini.
func (u *Usecase) Me(ctx context.Context, userID int64) (Profile, error) {
	user, err := u.store.GetUserDetailByID(ctx, userID)
	if err != nil {
		return Profile{}, apperr.NotFound(err, apperr.Unauthenticated)
	}
	if reason := accessError(user); reason != "" {
		return Profile{}, reasonError(reason)
	}
	permissions, err := permissionsOf(ctx, u.store.Queries, user)
	return Profile{User: user, Permissions: permissions}, err
}
