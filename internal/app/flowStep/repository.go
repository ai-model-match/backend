package flowStep

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type flowStepRepositoryInterface interface {
	checkFlowExists(tx *gorm.DB, flowID uuid.UUID) (bool, error)
	checkUseCaseStepExists(tx *gorm.DB, useCaseStepID uuid.UUID) (bool, error)
	listFlowSteps(tx *gorm.DB, flowID uuid.UUID, limit int, offset int, forUpdate bool) ([]flowStepEntity, int64, error)
	getFlowStepByID(tx *gorm.DB, flowStepID uuid.UUID, forUpdate bool) (flowStepEntity, error)
	saveFlowStep(tx *gorm.DB, flowStep flowStepEntity, operation mm_db.SaveOperation) (flowStepEntity, error)
	getAllMissingFlowSteps(tx *gorm.DB, useCaseID uuid.UUID) ([]missingFlowStepEntity, error)
}

type flowStepRepository struct {
}

func newFlowStepRepository() flowStepRepository {
	return flowStepRepository{}
}

func (r flowStepRepository) checkFlowExists(tx *gorm.DB, flowID uuid.UUID) (bool, error) {
	var model *flowModel
	query := tx.Where("id = ?", flowID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return false, nil
	}
	return true, nil
}

func (r flowStepRepository) checkUseCaseStepExists(tx *gorm.DB, useCaseStepID uuid.UUID) (bool, error) {
	var model *useCaseStepModel
	query := tx.Where("id = ?", useCaseStepID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return false, nil
	}
	return true, nil
}

func (r flowStepRepository) listFlowSteps(tx *gorm.DB, flowID uuid.UUID, limit int, offset int, forUpdate bool) ([]flowStepEntity, int64, error) {
	var totalCount int64
	// Retrieve Flow Steps keeping the order of Use Case Steps
	var models []*flowStepModel
	query := tx.Model(&flowStepModel{}).
		Select("mm_flow_step.*").
		Joins("JOIN mm_use_case_step us ON mm_flow_step.use_case_step_id = us.id").
		Where("mm_flow_step.flow_id = ?", flowID).
		Order("us.position ASC").
		Limit(limit).
		Offset(offset)

	queryCount := tx.Model(flowStepModel{}).Where("flow_id = ?", flowID)

	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(limit).Offset(offset).Find(&models)
	queryCount.Count(&totalCount)

	if result.Error != nil {
		return []flowStepEntity{}, 0, result.Error
	}
	var entities []flowStepEntity = []flowStepEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, totalCount, nil
}

func (r flowStepRepository) getFlowStepByID(tx *gorm.DB, flowStepID uuid.UUID, forUpdate bool) (flowStepEntity, error) {
	var model *flowStepModel
	query := tx.Where("id = ?", flowStepID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return flowStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowStepRepository) saveFlowStep(tx *gorm.DB, flowStep flowStepEntity, operation mm_db.SaveOperation) (flowStepEntity, error) {
	var model = flowStepModel(flowStep)
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
		return flowStepEntity{}, err
	}
	return flowStep, nil
}

func (r flowStepRepository) getAllMissingFlowSteps(tx *gorm.DB, useCaseID uuid.UUID) ([]missingFlowStepEntity, error) {
	var models []missingFlowStepModel
	query := `
		SELECT 
			f.id AS flow_id,
			f.use_case_id AS use_case_id,
			ucs.id AS use_case_step_id
		FROM mm_flow f
		CROSS JOIN mm_use_case_step ucs
		LEFT JOIN mm_flow_step fs
			ON fs.flow_id = f.id
			AND fs.use_case_step_id = ucs.id
		WHERE f.use_case_id = ?
		AND ucs.use_case_id = ?
		AND fs.id IS NULL
		ORDER BY ucs.position ASC
	`
	if err := tx.Raw(query, useCaseID, useCaseID).Scan(&models).Error; err != nil {
		return nil, err
	}
	var entities []missingFlowStepEntity = []missingFlowStepEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, nil
}
