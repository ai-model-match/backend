package rsEngine

import "gorm.io/gorm"

type rsEngineRepositoryInterface interface {
	getRolloutStrategyByUseCaseIDAndStatus(tx *gorm.DB, useCaseID string, status RolloutState) (rolloutStrategyEntity, error)
}

type rsEngineRepository struct {
}

func newRsEngineRepository() rsEngineRepository {
	return rsEngineRepository{}
}

func (r rsEngineRepository) getRolloutStrategyByUseCaseIDAndStatus(tx *gorm.DB, useCaseID string, status RolloutState) (rolloutStrategyEntity, error) {
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
