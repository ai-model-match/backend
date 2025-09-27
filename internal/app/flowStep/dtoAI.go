package flowStep

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type aiRequestDTO struct {
	Modality   string           `json:"modality"`
	Parameters *json.RawMessage `json:"parameters,omitempty"`
}

func (r aiRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Modality, validation.Required),
		validation.Field(&r.Parameters, validation.Required),
	)
}
