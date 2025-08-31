package rolloutStrategy

import (
	"encoding/json"
	"slices"
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
	var rolloutStrategy rolloutStrategyEntity
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
		config, _ := json.Marshal(map[string]interface{}{})
		rolloutStrategy = rolloutStrategyEntity{
			ID:            uuid.New(),
			UseCaseID:     useCaseID,
			RolloutState:  RolloutStateInit,
			Configuration: json.RawMessage(config),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if _, err := s.repository.saveRolloutStrategy(tx, rolloutStrategy, mm_db.Create); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return rolloutStrategyEntity{}, errTransaction
	}
	return rolloutStrategy, nil
}

func (s rolloutStrategyService) updateRolloutStrategy(ctx *gin.Context, input updateRolloutStrategyInputDto) (rolloutStrategyEntity, error) {
	now := time.Now()
	var rolloutStrategy rolloutStrategyEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Check if the use Case exists
		useCaseID := uuid.MustParse(input.UseCaseID)
		if item, err := s.repository.getRolloutStrategyByUseCaseID(tx, useCaseID, true); err != nil {
			return mm_err.ErrGeneric
		} else if mm_utils.IsEmpty(item) {
			return errRolloutStrategyNotFound
		} else {
			rolloutStrategy = item
		}
		// Check request to change Rollout state
		if input.RolloutState != nil && RolloutState(*input.RolloutState) != rolloutStrategy.RolloutState {
			// Check the flow, if cna be move to next state
			if ok := s.checkStateFlow(rolloutStrategy.RolloutState, RolloutState(*input.RolloutState)); ok {
				rolloutStrategy.RolloutState = RolloutState(*input.RolloutState)
			} else {
				return errRolloutStrategyTransitionStateNotAllowed
			}
		}
		// Check request to change Rollout configuration
		if input.Configuration != nil {
			// not allowed if the state is not INIT
			if rolloutStrategy.RolloutState != RolloutStateInit {
				return errRolloutStrategyTransitionStateNotAllowed
			}
			// Round decimals on percentages
			for i := range input.Configuration.Warmup.Goals {
				input.Configuration.Warmup.Goals[i].FinalServePct = *(mm_utils.RoundTo2Decimals(&input.Configuration.Warmup.Goals[i].FinalServePct))
			}
			// Finally, convert in JSON
			if configuration, err := json.Marshal(input.Configuration); err != nil {
				return errRolloutStrategyWrongConfigFormat
			} else {
				rolloutStrategy.Configuration = configuration
			}
		}
		// Save Rollout Strategy
		rolloutStrategy.UpdatedAt = now
		if _, err := s.repository.saveRolloutStrategy(tx, rolloutStrategy, mm_db.Update); err != nil {
			return mm_err.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return rolloutStrategyEntity{}, errTransaction
	}
	return rolloutStrategy, nil
}

func (s rolloutStrategyService) checkStateFlow(currentState RolloutState, nextState RolloutState) bool {
	if nextStates, ok := allowedTransitions[currentState]; ok {
		if slices.Contains(nextStates, nextState) {
			return true
		}
	}
	return false
}
