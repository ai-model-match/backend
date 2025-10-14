package flowStep

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type ListFlowStepsInputDto struct {
	FlowID   string `form:"flowId"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

func (r ListFlowStepsInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.Page, validation.Required, validation.Min(1)),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1), validation.Max(200)),
	)
}

type getFlowStepInputDto struct {
	ID string `uri:"flowStepId"`
}

func (r getFlowStepInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
	)
}

type updateFlowStepInputDto struct {
	ID            string       `uri:"flowStepId"`
	Configuration aiRequestDTO `json:"configuration"`
}

func (r updateFlowStepInputDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, is.UUID),
		validation.Field(&r.Configuration, validation.Required, validation.By(func(v interface{}) error {
			return v.(aiRequestDTO).validate()
		})),
	)
}

func getDefaultConfiguration() json.RawMessage {
	return json.RawMessage([]byte(`{
  "modality": "chat.completions",
  "parameters": {
    "model": "gpt-5",
    "temperature": 0.7,
    "max_tokens": 100,
    "frequency_penalty": 0,
    "presence_penalty": 0,
    "n": 1,
    "top_p": 1,
    "stream": false,
    "messages": [
      {
        "role": "system",
        "content": "You are an AI assistant that provides concise and clear responses."
      },
      {
        "role": "user",
        "content": "<<USER_INPUT>>"
      }
    ],
    "user": "<<CORRELATION_ID>>"
  }
}`))
}

func getDefaultPlaceholders() json.RawMessage {
	params, _ := json.Marshal([]string{
		"USER_INPUT",
		"CORRELATION_ID",
	})
	return json.RawMessage(params)
}
