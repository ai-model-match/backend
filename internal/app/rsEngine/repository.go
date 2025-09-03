package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rsEngineRepositoryInterface interface {
	getRolloutStrategyByUseCaseIDAndStatus(tx *gorm.DB, useCaseID uuid.UUID, status mm_pubsub.RolloutState) (rolloutStrategyEntity, error)
	getFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID, forUpdate bool) ([]flowEntity, error)
	getFlowStatisticsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowStatisticsEntity, error)
	saveFlow(tx *gorm.DB, flow flowEntity, operation mm_db.SaveOperation) (flowEntity, error)
	saveRolloutStrategy(tx *gorm.DB, rolloutStrategy rolloutStrategyEntity, operation mm_db.SaveOperation) (rolloutStrategyEntity, error)
}

type rsEngineRepository struct {
}

func newRsEngineRepository() rsEngineRepository {
	return rsEngineRepository{}
}

func (r rsEngineRepository) getRolloutStrategyByUseCaseIDAndStatus(tx *gorm.DB, useCaseID uuid.UUID, status mm_pubsub.RolloutState) (rolloutStrategyEntity, error) {
	var model *rolloutStrategyModel
	query := tx.Where("use_case_id = ? AND rollout_state = ?", useCaseID, status)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return rolloutStrategyEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return rolloutStrategyEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r rsEngineRepository) getFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID, forUpdate bool) ([]flowEntity, error) {
	var models []flowModel
	query := tx.Model(flowModel{}).Where("use_case_id = ?", useCaseID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
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

func (r rsEngineRepository) saveFlow(tx *gorm.DB, flow flowEntity, operation mm_db.SaveOperation) (flowEntity, error) {
	var model = flowModel(flow)
	var err error
	switch operation {
	case mm_db.Create:
		err = tx.Create(model).Error
	case mm_db.Update:
		err = tx.Updates(model).Error
	case mm_db.Upsert:
		err = tx.Save(model).Error
	}
	if err != nil {
		return flowEntity{}, err
	}
	return flow, nil
}

func (r rsEngineRepository) saveRolloutStrategy(tx *gorm.DB, rolloutStrategy rolloutStrategyEntity, operation mm_db.SaveOperation) (rolloutStrategyEntity, error) {
	var err error
	var model rolloutStrategyModel
	if err = model.fromEntity(rolloutStrategy); err != nil {
		return rolloutStrategyEntity{}, err
	}
	switch operation {
	case mm_db.Create:
		err = tx.Create(model).Error
	case mm_db.Update:
		err = tx.Updates(model).Error
	case mm_db.Upsert:
		err = tx.Save(model).Error
	}
	if err != nil {
		return rolloutStrategyEntity{}, err
	}
	return rolloutStrategy, nil
}
