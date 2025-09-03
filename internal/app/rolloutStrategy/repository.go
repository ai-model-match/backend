package rolloutStrategy

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rolloutStrategyRepositoryInterface interface {
	checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error)
	getRolloutStrategyByUseCaseID(tx *gorm.DB, flowIDuseCaseID uuid.UUID, forUpdate bool) (rolloutStrategyEntity, error)
	saveRolloutStrategy(tx *gorm.DB, rolloutStrategy rolloutStrategyEntity, operation mm_db.SaveOperation) (rolloutStrategyEntity, error)
}

type rolloutStrategyRepository struct {
}

func newRolloutStrategyRepository() rolloutStrategyRepository {
	return rolloutStrategyRepository{}
}

func (r rolloutStrategyRepository) checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error) {
	var model *useCaseModel
	query := tx.Where("id = ?", useCaseID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return false, nil
	}
	return true, nil
}
func (r rolloutStrategyRepository) getRolloutStrategyByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID, forUpdate bool) (rolloutStrategyEntity, error) {
	var model *rolloutStrategyModel
	query := tx.Where("use_case_id = ?", useCaseID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return rolloutStrategyEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return rolloutStrategyEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r rolloutStrategyRepository) saveRolloutStrategy(tx *gorm.DB, rolloutStrategy rolloutStrategyEntity, operation mm_db.SaveOperation) (rolloutStrategyEntity, error) {
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
