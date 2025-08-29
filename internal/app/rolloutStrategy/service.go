package rolloutStrategy

import (
	"encoding/json"
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
	getRolloutStrategyByID(ctx *gin.Context, input getRolloutStrategyInputDto) (rolloutStrategyEntity, error)
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

func (s rolloutStrategyService) getRolloutStrategyByID(ctx *gin.Context, input getRolloutStrategyInputDto) (rolloutStrategyEntity, error) {
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
		// Retrieve the Flow Step and check if exists
		useCase, err := s.repository.getUseCaseByID(tx, useCaseID)
		if err != nil {
			return mm_err.ErrGeneric
		}
		if mm_utils.IsEmpty(useCase) {
			return errUseCaseNotFound
		}
		// Check if the Rollout Strategy already exists, if yes, return an error
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
			UseCaseID:     useCase.ID,
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
		if configuration, err := json.Marshal(input.Configuration); err != nil {
			return errRolloutStrategyWrongConfigFormat
		} else {
			rolloutStrategy.Configuration = configuration
		}
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
