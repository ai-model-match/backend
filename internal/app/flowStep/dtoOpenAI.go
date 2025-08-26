package flowStep

import (
	"fmt"

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
	Modality modalityType `json:"modality"`

	// Sub-structs for each modality
	Chat     *chatCompletionRequestDTO     `json:"chat,omitempty"`
	Image    *imageGenerationRequestDTO    `json:"image,omitempty"`
	Audio    *audioTranscriptionRequestDTO `json:"audio,omitempty"`
	File     *fileUploadRequestDTO         `json:"file,omitempty"`
	FineTune *fineTuneRequestDTO           `json:"fine_tune,omitempty"`
}

// Validate validates the openAIRequestDTO based on modality
func (r openAIRequestDTO) validate() error {
	// Top-level validation
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Modality, validation.Required, validation.In(
			ModalityChat, ModalityImage, ModalityAudio, ModalityFile, ModalityFineTune,
		)),
	); err != nil {
		return err
	}

	// Validate sub-struct based on modality
	switch r.Modality {
	case ModalityChat:
		if r.Chat == nil {
			return fmt.Errorf("chat data is required for modality 'chat'")
		}
		return r.Chat.validate()
	case ModalityImage:
		if r.Image == nil {
			return fmt.Errorf("image data is required for modality 'image'")
		}
		return r.Image.validate()
	case ModalityAudio:
		if r.Audio == nil {
			return fmt.Errorf("audio data is required for modality 'audio'")
		}
		return r.Audio.validate()
	case ModalityFile:
		if r.File == nil {
			return fmt.Errorf("file data is required for modality 'file'")
		}
		return r.File.validate()
	case ModalityFineTune:
		if r.FineTune == nil {
			return fmt.Errorf("fine-tune data is required for modality 'fine-tune'")
		}
		return r.FineTune.validate()
	default:
		return fmt.Errorf("unsupported modality: %s", r.Modality)
	}
}

// ==============================
// Chat DTO
// ==============================

type chatCompletionRequestDTO struct {
	Model            string       `json:"model"`
	Messages         []messageDTO `json:"messages"`
	Temperature      *float32     `json:"temperature,omitempty"`
	TopP             *float32     `json:"top_p,omitempty"`
	N                *int         `json:"n,omitempty"`
	Stream           *bool        `json:"stream,omitempty"`
	Stop             []string     `json:"stop,omitempty"`
	MaxTokens        *int         `json:"max_tokens,omitempty"`
	PresencePenalty  *float32     `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32     `json:"frequency_penalty,omitempty"`
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
		validation.Field(&r.Prompt, validation.Required, validation.Length(1, 1000)),
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
	Temperature *float32 `json:"temperature,omitempty"`
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
