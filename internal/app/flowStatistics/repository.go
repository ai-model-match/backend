package flowStatistics

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type flowStatisticsRepositoryInterface interface {
	getFlowByID(tx *gorm.DB, flowID uuid.UUID) (flowEntity, error)
	getFlowStatisticsByFlowID(tx *gorm.DB, flowID uuid.UUID, forUpdate bool) (flowStatisticsEntity, error)
	saveFlowStatistics(tx *gorm.DB, flowStatistics flowStatisticsEntity, operation mm_db.SaveOperation) (flowStatisticsEntity, error)
}

type flowStatisticsRepository struct {
}

func newFlowStatisticsRepository() flowStatisticsRepository {
	return flowStatisticsRepository{}
}

func (r flowStatisticsRepository) getFlowByID(tx *gorm.DB, flowID uuid.UUID) (flowEntity, error) {
	var model *flowModel
	query := tx.Where("id = ?", flowID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return flowEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowStatisticsRepository) getFlowStatisticsByFlowID(tx *gorm.DB, flowID uuid.UUID, forUpdate bool) (flowStatisticsEntity, error) {
	var model *flowStatisticsModel
	query := tx.Where("flow_id = ?", flowID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowStatisticsEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return flowStatisticsEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowStatisticsRepository) saveFlowStatistics(tx *gorm.DB, flowStatistics flowStatisticsEntity, operation mm_db.SaveOperation) (flowStatisticsEntity, error) {
	var model = flowStatisticsModel(flowStatistics)
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
		return flowStatisticsEntity{}, err
	}
	return flowStatistics, nil
}
