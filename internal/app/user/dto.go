package user

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type getUserInputDto struct {
	id string `uri:"userID"`
}

func (r getUserInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.id, validation.Required, is.UUID),
	)
}

type createUserInputDto struct {
	ID        string
	Firstname string
	Lastname  string
	Email     string
}

func (r createUserInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Firstname, validation.Required, validation.Length(3, 255)),
		validation.Field(&r.Lastname, validation.Required, validation.Length(3, 255)),
		validation.Field(&r.Email, validation.Required, is.Email),
	)
}
