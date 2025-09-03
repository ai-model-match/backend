package rolloutStrategy

import (
	"errors"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rsEscapePhaseDto struct {
	Rules []rsEscapeRuleDto `json:"rules"`
}

func (r rsEscapePhaseDto) validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.Rules, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
			v := value.(rsEscapeRuleDto)
			return v.validate()
		}))),
	); err != nil {
		return err
	}
	// Check there is only one rule per Flow
	seen := make(map[string]bool)
	for _, rule := range r.Rules {
		if _, exists := seen[rule.FlowID]; exists {
			return errors.New("flow can have only one rule associated")
		}
		seen[rule.FlowID] = true
	}
	return nil
}

func (r rsEscapePhaseDto) toEntity() mm_pubsub.RsEscapePhase {
	rules := make([]mm_pubsub.RsEscapeRule, len(r.Rules))
	for i, rule := range r.Rules {
		rules[i] = rule.toEntity()
	}
	return mm_pubsub.RsEscapePhase{
		Rules: rules,
	}
}

type rsEscapeRuleDto struct {
	FlowID      string                `json:"flow_id"`
	MinFeedback int64                 `json:"min_feedback"`
	LowerScore  *float64              `json:"lower_score"`
	Rollback    []rsEscapeRollbackDto `json:"rollback"`
}

func (r rsEscapeRuleDto) validate() error {
	if err := validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.MinFeedback, validation.Required, validation.Min(int64(1))),
		validation.Field(&r.LowerScore, validation.Required, validation.Min(MinFeedbackScore), validation.Max(MaxFeedbackScore)),
		validation.Field(&r.Rollback, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value interface{}) error {
			v := value.(rsEscapeRollbackDto)
			return v.validate()
		}))),
	); err != nil {
		return err
	}
	// Check there is only one goal per Flow
	seen := make(map[string]bool)
	totPct := float64(0)
	for _, rollback := range r.Rollback {
		totPct = totPct + *mm_utils.RoundTo2DecimalsPtr(rollback.FinalServePct)
		if _, exists := seen[rollback.FlowID]; exists {
			return errors.New("flow can have only one rollback associated")
		}
		seen[rollback.FlowID] = true
	}
	if totPct > 100 {
		return errors.New("rollbacks must have a maximum of 100% as sum")
	}
	return nil
}

func (r rsEscapeRuleDto) toEntity() mm_pubsub.RsEscapeRule {
	rollbacks := []mm_pubsub.RsEscapeRollback{}
	for _, rollback := range r.Rollback {
		rollbacks = append(rollbacks, rollback.toEntity())
	}
	return mm_pubsub.RsEscapeRule{
		FlowID:      r.FlowID,
		MinFeedback: r.MinFeedback,
		LowerScore:  *r.LowerScore,
		Rollback:    rollbacks,
	}
}

type rsEscapeRollbackDto struct {
	FlowID        string   `json:"flow_id"`
	FinalServePct *float64 `json:"final_serve_pct"`
}

func (r rsEscapeRollbackDto) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.FlowID, validation.Required, is.UUID),
		validation.Field(&r.FinalServePct, validation.When(r.FinalServePct != nil, validation.Min(float64(0)), validation.Max(float64(100)))),
	)
}

func (r rsEscapeRollbackDto) toEntity() mm_pubsub.RsEscapeRollback {
	return mm_pubsub.RsEscapeRollback{
		FlowID:        r.FlowID,
		FinalServePct: *r.FinalServePct,
	}
}
