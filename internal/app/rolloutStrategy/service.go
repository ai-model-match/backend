package rolloutStrategy

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_err"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rolloutStrategyServiceInterface interface {
	getRolloutStrategyByUseCaseID(ctx *gin.Context, input getRolloutStrategyInputDto) (rolloutStrategyEntity, error)
	createRolloutStrategy(useCaseID uuid.UUID) (rolloutStrategyEntity, error)
	updateRolloutStrategy(ctx *gin.Context, input updateRolloutStrategyInputDto) (rolloutStrategyEntity, error)
}

type rolloutStrategyService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  rolloutStrategyRepositoryInterface
}

func newRolloutStrategyService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository rolloutStrategyRepositoryInterface) rolloutStrategyService {
	return rolloutStrategyService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s rolloutStrategyService) getRolloutStrategyByUseCaseID(ctx *gin.Context, input getRolloutStrategyInputDto) (rolloutStrategyEntity, error) {
	useCaseID := uuid.MustParse(input.UseCaseID)
	item, err := s.repository.getRolloutStrategyByUseCaseID(s.storage, useCaseID, false)
	if err != nil {
		return rolloutStrategyEntity{}, mm_err.ErrGeneric
	}
	if mm_utils.IsEmpty(item) {
		return rolloutStrategyEntity{}, errRolloutStrategyNotFound
	}
	return item, nil
}

func (s rolloutStrategyService) createRolloutStrategy(useCaseID uuid.UUID) (rolloutStrategyEntity, error) {
	now := time.Now()
	var newRolloutStrategy rolloutStrategyEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Retrieve and check if the related Use Case exists
		exists, err := s.repository.checkUseCaseExists(tx, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !exists {
			return errUseCaseNotFound
		}
		// Check if the Rollout Strategy already exists
		item, err := s.repository.getRolloutStrategyByUseCaseID(tx, useCaseID, false)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if !mm_utils.IsEmpty(item) {
			return errRolloutStrategyAlreadyExists
		}
		// Create the new Rollout Strategy with default values and store it
		newRolloutStrategy = rolloutStrategyEntity{
			ID:           uuid.New(),
			UseCaseID:    useCaseID,
			RolloutState: mm_pubsub.RolloutStateInit,
			Configuration: mm_pubsub.RSConfiguration{
				Warmup: nil,
				Escape: nil,
				Adaptive: mm_pubsub.RsAdaptivePhase{
					MinFeedback:  0,
					MaxStepPct:   10,
					IntervalMins: 10,
				},
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := s.repository.saveRolloutStrategy(tx, newRolloutStrategy, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of Rollout Straregy created
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicRolloutStrategyV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.RolloutStrategyCreatedEvent,
				EventEntity: &mm_pubsub.RolloutStrategyEventEntity{
					ID:            newRolloutStrategy.ID,
					UseCaseID:     newRolloutStrategy.UseCaseID,
					RolloutState:  newRolloutStrategy.RolloutState,
					Configuration: newRolloutStrategy.Configuration,
					CreatedAt:     newRolloutStrategy.CreatedAt,
					UpdatedAt:     newRolloutStrategy.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(rolloutStrategyEntity{}, newRolloutStrategy),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return rolloutStrategyEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return newRolloutStrategy, nil
}

func (s rolloutStrategyService) updateRolloutStrategy(ctx *gin.Context, input updateRolloutStrategyInputDto) (rolloutStrategyEntity, error) {
	now := time.Now()
	var updatedRolloutStrategy rolloutStrategyEntity
	eventsToPublish := []mm_pubsub.EventToPublish{}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseID := uuid.MustParse(input.UseCaseID)
		currentRolloutStrategy, err := s.repository.getRolloutStrategyByUseCaseID(tx, useCaseID, true)
		if err != nil {
			return mm_err.ErrGeneric
		} else if mm_utils.IsEmpty(currentRolloutStrategy) {
			return errRolloutStrategyNotFound
		} else {
			updatedRolloutStrategy = currentRolloutStrategy
		}
		// Avoid change configuration with Rollout State different from INIT
		if input.Configuration != nil && updatedRolloutStrategy.RolloutState != mm_pubsub.RolloutStateInit {
			return errRolloutStrategyNotEditableWhileActive
		}
		// Check request to change Rollout state
		if input.RolloutState != nil {
			// Check the flow, if cna be move to next state
			if ok := checkStateFlow(updatedRolloutStrategy.RolloutState, mm_pubsub.RolloutState(*input.RolloutState)); ok {
				updatedRolloutStrategy.RolloutState = mm_pubsub.RolloutState(*input.RolloutState)
			} else {
				return errRolloutStrategyTransitionStateNotAllowed
			}
			// Now, if we are activating Rollout Strategy (from INIT to WARMUP), but there is no warmup config, move to ADAPT
			if updatedRolloutStrategy.RolloutState == mm_pubsub.RolloutStateWarmup && mm_utils.IsEmpty(updatedRolloutStrategy.Configuration.Warmup) {
				updatedRolloutStrategy.RolloutState = mm_pubsub.RolloutStateAdaptive
			}
		}
		// Check request to change Rollout configuration
		if input.Configuration != nil {
			// not allowed if the state is not INIT
			if updatedRolloutStrategy.RolloutState != mm_pubsub.RolloutStateInit {
				return errRolloutStrategyTransitionStateNotAllowed
			}
			// Round decimals on percentages for Warmup
			if !mm_utils.IsEmpty(input.Configuration.Warmup) {
				for i := range input.Configuration.Warmup.Goals {
					input.Configuration.Warmup.Goals[i].FinalServePct = (mm_utils.RoundTo2DecimalsPtr(input.Configuration.Warmup.Goals[i].FinalServePct))
				}
			}
			// Round decimals on percentages for Escape
			if !mm_utils.IsEmpty(input.Configuration.Escape) {
				for i := range input.Configuration.Escape.Rules {
					input.Configuration.Escape.Rules[i].LowerScore = (mm_utils.RoundTo2DecimalsPtr(input.Configuration.Escape.Rules[i].LowerScore))
					for j := range input.Configuration.Escape.Rules[i].Rollback {
						input.Configuration.Escape.Rules[i].Rollback[j].FinalServePct = (mm_utils.RoundTo2DecimalsPtr(input.Configuration.Escape.Rules[i].Rollback[j].FinalServePct))
					}
				}
			}
			// Round decimals on percentages for Adaptive
			input.Configuration.Adaptive.MaxStepPct = (mm_utils.RoundTo2Decimals(input.Configuration.Adaptive.MaxStepPct))

			// Update the configuration
			updatedRolloutStrategy.Configuration = input.Configuration.toEntity()
		}
		// Save Rollout Strategy
		updatedRolloutStrategy.UpdatedAt = now
		if _, err := s.repository.saveRolloutStrategy(tx, updatedRolloutStrategy, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		}
		// Send an event of Rollout Straregy updated
		if event, err := s.pubSubAgent.Persist(tx, mm_pubsub.TopicRolloutStrategyV1, mm_pubsub.PubSubMessage{
			Message: mm_pubsub.PubSubEvent{
				EventID:   uuid.New(),
				EventTime: time.Now(),
				EventType: mm_pubsub.RolloutStrategyUpdatedEvent,
				EventEntity: &mm_pubsub.RolloutStrategyEventEntity{
					ID:            updatedRolloutStrategy.ID,
					UseCaseID:     updatedRolloutStrategy.UseCaseID,
					RolloutState:  updatedRolloutStrategy.RolloutState,
					Configuration: updatedRolloutStrategy.Configuration,
					CreatedAt:     updatedRolloutStrategy.CreatedAt,
					UpdatedAt:     updatedRolloutStrategy.UpdatedAt,
				},
				EventChangedFields: mm_utils.DiffStructs(currentRolloutStrategy, updatedRolloutStrategy),
			},
		}); err != nil {
			return err
		} else {
			eventsToPublish = append(eventsToPublish, event)
		}
		return nil
	})
	if errTransaction != nil {
		return rolloutStrategyEntity{}, errTransaction
	} else {
		s.pubSubAgent.PublishBulk(eventsToPublish)
	}
	return updatedRolloutStrategy, nil
}
