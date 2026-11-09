package auth_usecase

import (
	"ppl-m1-be/internal/usecase/repository"
	"ppl-m1-be/internal/usecase/utils"
)

type Usecase struct {
	store     *repository.Store
	tokens    *utils.TokenMaker
	dummyHash string
}

func New(store *repository.Store, tokens *utils.TokenMaker) *Usecase {
	return &Usecase{
		store:     store,
		tokens:    tokens,
		dummyHash: newDummyHash(),
	}
}

// newDummyHash dipakai saat email tidak terdaftar agar waktu respons sama
// dengan saat kata sandi salah.
func newDummyHash() string {
	hash, err := utils.HashPassword("dummy-password-for-timing")
	if err != nil {
		panic(err)
	}
	return hash
}
