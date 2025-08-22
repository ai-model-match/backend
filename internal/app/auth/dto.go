package auth

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type loginUserInputDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r loginUserInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Username, validation.Required),
		validation.Field(&r.Password, validation.Required),
	)
}

type refreshTokenInputDto struct {
	RefreshToken string `json:"refreshToken"`
}

func (r refreshTokenInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RefreshToken, validation.Required),
	)
}
