package flowStepStatistics

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type flowStepStatisticsRepositoryInterface interface {
	getFlowStepByID(tx *gorm.DB, flowStepID uuid.UUID) (flowStepEntity, error)
	getFlowStepStatisticsByFlowStepID(tx *gorm.DB, flowStepID uuid.UUID, forUpdate bool) (flowStepStatisticsEntity, error)
	saveFlowStepStatistics(tx *gorm.DB, flowStepStatistics flowStepStatisticsEntity, operation mm_db.SaveOperation) (flowStepStatisticsEntity, error)
	cleanupFlowStepStatisticsByUseCaseId(tx *gorm.DB, useCaseID uuid.UUID) error
}

type flowStepStatisticsRepository struct {
}

func newFlowStepStatisticsRepository() flowStepStatisticsRepository {
	return flowStepStatisticsRepository{}
}

func (r flowStepStatisticsRepository) getFlowStepByID(tx *gorm.DB, flowID uuid.UUID) (flowStepEntity, error) {
	var model *flowStepModel
	query := tx.Where("id = ?", flowID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return flowStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowStepStatisticsRepository) getFlowStepStatisticsByFlowStepID(tx *gorm.DB, flowStepID uuid.UUID, forUpdate bool) (flowStepStatisticsEntity, error) {
	var model *flowStepStatisticsModel
	query := tx.Where("flow_step_id = ?", flowStepID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowStepStatisticsEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return flowStepStatisticsEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowStepStatisticsRepository) saveFlowStepStatistics(tx *gorm.DB, flowStepStatistics flowStepStatisticsEntity, operation mm_db.SaveOperation) (flowStepStatisticsEntity, error) {
	var model = flowStepStatisticsModel(flowStepStatistics)
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
		return flowStepStatisticsEntity{}, err
	}
	return flowStepStatistics, nil
}

func (r flowStepStatisticsRepository) cleanupFlowStepStatisticsByUseCaseId(tx *gorm.DB, useCaseID uuid.UUID) error {
	result := tx.Model(&flowStepStatisticsModel{}).
		Where("flow_step_id IN (?)",
			tx.Model(&flowStepModel{}).Select("id").Where("use_case_id = ?", useCaseID),
		).
		Update("tot_req", 0)
	return result.Error
}
