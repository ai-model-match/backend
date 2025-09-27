package flowStep

import (
	"encoding/json"
	"errors"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// ==============================
// Modality
// ==============================

type modalityType string

const (
	ModalityChat     modalityType = "chat"
	ModalityImage    modalityType = "image"
	ModalityAudio    modalityType = "audio"
	ModalityFile     modalityType = "file"
	ModalityFineTune modalityType = "fine-tune"
)

// ==============================
// Unified OpenAI Request DTO
// ==============================

type openAIRequestDTO struct {
	Modality   modalityType     `json:"modality"`
	Parameters *json.RawMessage `json:"parameters,omitempty"`
}

var modalityTypes = map[modalityType]reflect.Type{
	ModalityChat:     reflect.TypeOf(chatCompletionRequestDTO{}),
	ModalityImage:    reflect.TypeOf(imageGenerationRequestDTO{}),
	ModalityAudio:    reflect.TypeOf(audioTranscriptionRequestDTO{}),
	ModalityFile:     reflect.TypeOf(fileUploadRequestDTO{}),
	ModalityFineTune: reflect.TypeOf(fineTuneRequestDTO{}),
}

// Validate validates the openAIRequestDTO based on modality
func (r openAIRequestDTO) validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Modality, validation.Required, validation.In(
			ModalityChat, ModalityImage, ModalityAudio, ModalityFile, ModalityFineTune,
		)),
		validation.Field(&r.Parameters, validation.Required),
	); err != nil {
		return err
	}
	cfg, err := r.ParseParameters()
	if err != nil {
		return err
	}
	return cfg.validate()
}

type validatable interface {
	validate() error
}

func (r *openAIRequestDTO) ParseParameters() (validatable, error) {
	dtoType, ok := modalityTypes[r.Modality]
	if !ok {
		return nil, errors.New("mismatched modality and configuration: " + string(r.Modality))
	}
	cfgPtr := reflect.New(dtoType).Interface()
	if err := json.Unmarshal(*r.Parameters, cfgPtr); err != nil {
		return nil, err
	}
	return cfgPtr.(validatable), nil
}

// ==============================
// Chat DTO
// ==============================

type chatCompletionRequestDTO struct {
	Model            string       `json:"model"`
	Messages         []messageDTO `json:"messages"`
	Temperature      *float64     `json:"temperature,omitempty"`
	TopP             *float64     `json:"top_p,omitempty"`
	N                *int         `json:"n,omitempty"`
	Stream           *bool        `json:"stream,omitempty"`
	Stop             []string     `json:"stop,omitempty"`
	MaxTokens        *int         `json:"max_tokens,omitempty"`
	PresencePenalty  *float64     `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64     `json:"frequency_penalty,omitempty"`
	User             *string      `json:"user,omitempty"`
}

type messageDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Validate chatCompletionRequestDTO
func (r chatCompletionRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Model, validation.Required),
		validation.Field(&r.Messages, validation.Required, validation.Length(1, 0)),
		validation.Field(&r.Stop, validation.Length(0, 4)),
	)
}

// ==============================
// Image DTO
// ==============================

type imageGenerationRequestDTO struct {
	Model   string  `json:"model,omitempty"`
	Prompt  string  `json:"prompt"`
	N       *int    `json:"n,omitempty"`
	Size    *string `json:"size,omitempty"`
	Quality *string `json:"quality,omitempty"`
	Style   *string `json:"style,omitempty"`
	User    *string `json:"user,omitempty"`
}

func (r imageGenerationRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Prompt, validation.Required, validation.Length(1, 0)),
	)
}

// ==============================
// Audio DTO
// ==============================

type audioTranscriptionRequestDTO struct {
	Model       string   `json:"model"`
	FilePath    string   `json:"file_path"`
	Prompt      *string  `json:"prompt,omitempty"`
	ResponseFmt *string  `json:"response_format,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Language    *string  `json:"language,omitempty"`
}

func (r audioTranscriptionRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Model, validation.Required),
		validation.Field(&r.FilePath, validation.Required),
	)
}

// ==============================
// File DTO
// ==============================

type fileUploadRequestDTO struct {
	FilePath string `json:"file_path"`
	Purpose  string `json:"purpose"`
}

func (r fileUploadRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FilePath, validation.Required),
		validation.Field(&r.Purpose, validation.Required),
	)
}

// ==============================
// Fine-Tune DTO
// ==============================

type fineTuneRequestDTO struct {
	Model        string  `json:"model"`
	TrainingFile string  `json:"training_file"`
	Suffix       *string `json:"suffix,omitempty"`
}

func (r fineTuneRequestDTO) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Model, validation.Required),
		validation.Field(&r.TrainingFile, validation.Required),
	)
}
