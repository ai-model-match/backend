package picker

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pickerRepositoryInterface interface {
	getUseCaseByCode(tx *gorm.DB, code string) (useCaseEntity, error)
	getUseCaseStepByCode(tx *gorm.DB, useCaseID uuid.UUID, code string) (useCaseStepEntity, error)
	getRecentCorrelationByID(tx *gorm.DB, correlationID uuid.UUID) (pickerCorrelationEntity, error)
	getFlowByID(tx *gorm.DB, flowID uuid.UUID) (flowEntity, error)
	getFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowEntity, error)
	getFlowStepByFlowIdandUseCaseStepId(tx *gorm.DB, FlowID uuid.UUID, UseCaseStepID uuid.UUID) (flowStepEntity, error)
	saveCorrelation(tx *gorm.DB, correlation pickerCorrelationEntity, operation mm_db.SaveOperation) (pickerCorrelationEntity, error)
	cleanUpExpiredPickerCorrelations(tx *gorm.DB) error
}

type pickerRepository struct {
}

func newPickerRepository() pickerRepository {
	return pickerRepository{}
}

func (r pickerRepository) getUseCaseByCode(tx *gorm.DB, code string) (useCaseEntity, error) {
	var model *useCaseModel
	query := tx.Where("code = ?", code)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return useCaseEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r pickerRepository) getUseCaseStepByCode(tx *gorm.DB, useCaseID uuid.UUID, code string) (useCaseStepEntity, error) {
	var model *useCaseStepModel
	query := tx.Where("code = ?", code).Where("use_case_id = ?", useCaseID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return useCaseStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r pickerRepository) getRecentCorrelationByID(tx *gorm.DB, correlationID uuid.UUID) (pickerCorrelationEntity, error) {
	var model *pickerCorrelationModel
	query := tx.Where("id = ?", correlationID).Where("created_at >= NOW() - INTERVAL '24 hours'")
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return pickerCorrelationEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return pickerCorrelationEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r pickerRepository) getFlowsByUseCaseID(tx *gorm.DB, useCaseID uuid.UUID) ([]flowEntity, error) {
	var models []*flowModel
	query := tx.Model(flowModel{}).Where("use_case_id = ?", useCaseID)
	result := query.Find(&models)
	if result.Error != nil {
		return []flowEntity{}, result.Error
	}
	var entities []flowEntity = []flowEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, nil
}
func (r pickerRepository) getFlowByID(tx *gorm.DB, flowID uuid.UUID) (flowEntity, error) {
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

func (r pickerRepository) getFlowStepByFlowIdandUseCaseStepId(tx *gorm.DB, FlowID uuid.UUID, UseCaseStepID uuid.UUID) (flowStepEntity, error) {
	var model *flowStepModel
	query := tx.Where("flow_id = ?", FlowID).Where("use_case_step_id = ?", UseCaseStepID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return flowStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r pickerRepository) saveCorrelation(tx *gorm.DB, correlation pickerCorrelationEntity, operation mm_db.SaveOperation) (pickerCorrelationEntity, error) {
	var model = pickerCorrelationModel(correlation)
	var err error
	switch operation {
	case mm_db.Create:
		err = tx.Create(model).Error
	case mm_db.Update:
		err = tx.Updates(model).Error
	case mm_db.Upsert:
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(model).Error
	}
	if err != nil {
		return pickerCorrelationEntity{}, err
	}
	return correlation, nil
}
func (r pickerRepository) cleanUpExpiredPickerCorrelations(tx *gorm.DB) error {
	return tx.Where("created_at < NOW() - INTERVAL '24 hours'").Delete(&pickerCorrelationModel{}).Error
}
