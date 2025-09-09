package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rsEngineRepositoryInterface interface {
	getRolloutStrategyByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) (rolloutStrategyEntity, error)
	getActiveRolloutStrategiesInState(tx *gorm.DB, states []mm_pubsub.RolloutState) ([]rolloutStrategyEntity, error)
	getActiveFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowEntity, error)
	getFlowStatisticsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowStatisticsEntity, error)
}

type rsEngineRepository struct {
}

func newRsEngineRepository() rsEngineRepository {
	return rsEngineRepository{}
}

func (r rsEngineRepository) getRolloutStrategyByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) (rolloutStrategyEntity, error) {
	var model *rolloutStrategyModel
	query := tx.Where("use_case_id = ?", useCaseID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return rolloutStrategyEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return rolloutStrategyEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r rsEngineRepository) getActiveRolloutStrategiesInState(tx *gorm.DB, states []mm_pubsub.RolloutState) ([]rolloutStrategyEntity, error) {
	var models []rolloutStrategyModel
	query := tx.Model(rolloutStrategyModel{}).Where("rollout_state IN ?", states)
	result := query.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	entities := make([]rolloutStrategyEntity, len(models))
	for i, model := range models {
		entities[i] = model.toEntity()
	}
	return entities, nil
}

func (r rsEngineRepository) getActiveFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowEntity, error) {
	var models []flowModel
	query := tx.Model(flowModel{}).Where("use_case_id = ?", useCaseID).Where("active IS TRUE")
	result := query.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	entities := make([]flowEntity, len(models))
	for i, model := range models {
		entities[i] = model.toEntity()
	}
	return entities, nil
}

func (r rsEngineRepository) getFlowStatisticsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowStatisticsEntity, error) {
	var models []flowStatisticsModel
	query := tx.Model(flowStatisticsModel{}).Where("use_case_id = ?", useCaseID)

	result := query.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	entities := make([]flowStatisticsEntity, len(models))
	for i, model := range models {
		entities[i] = model.toEntity()
	}
	return entities, nil
}
