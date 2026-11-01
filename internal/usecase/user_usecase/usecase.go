package user_usecase

import "ppl-m1-be/internal/usecase/audit_usecase"

type Usecase struct {
	store userStore
	audit audit_usecase.Logger
}

func New(store userStore, audit audit_usecase.Logger) *Usecase {
	return &Usecase{store: store, audit: audit}
}