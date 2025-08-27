package flow

import (
	"fmt"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type flowRepositoryInterface interface {
	checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error)
	checkUseCaseIsActive(tx *gorm.DB, useCaseID uuid.UUID) (bool, error)
	listFlows(tx *gorm.DB, useCaseID uuid.UUID, limit int, offset int, orderBy flowOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]flowEntity, int64, error)
	getFlowByID(tx *gorm.DB, flowID uuid.UUID, forUpdate bool) (flowEntity, error)
	getFlowByCode(tx *gorm.DB, useCaseID uuid.UUID, flowCode string, forUpdate bool) (flowEntity, error)
	saveFlow(tx *gorm.DB, flow flowEntity, operation mm_db.SaveOperation) (flowEntity, error)
	deleteFlow(tx *gorm.DB, flow flowEntity) (flowEntity, error)
	makeFallbackConsistent(tx *gorm.DB, fallbackFlow flowEntity) error
}

type flowRepository struct {
	relevanceThresholdConfig float64
}

func newFlowRepository(relevanceThresholdConfig float64) flowRepository {
	return flowRepository{
		relevanceThresholdConfig: relevanceThresholdConfig,
	}
}

func (r flowRepository) checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error) {
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

func (r flowRepository) checkUseCaseIsActive(tx *gorm.DB, useCaseID uuid.UUID) (bool, error) {
	var model *useCaseModel
	query := tx.Where("id = ?", useCaseID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return false, errUseCaseNotFound
	}
	return model.Active, nil
}

func (r flowRepository) listFlows(tx *gorm.DB, useCaseID uuid.UUID, limit int, offset int, orderBy flowOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]flowEntity, int64, error) {
	var totalCount int64
	var order string

	var models []*flowModel
	query := tx.Model(flowModel{}).Where("use_case_id = ?", useCaseID)
	queryCount := tx.Model(flowModel{}).Where("use_case_id = ?", useCaseID)

	// The ordering of these fields is important for the relevance order
	searchFields := []string{"title", "description"}
	// Add fuzzy search query based on the provided search key and table fields
	if searchKey != nil {
		mm_db.GenerateFuzzySearch(query, *searchKey, searchFields, r.relevanceThresholdConfig)
		mm_db.GenerateFuzzySearch(queryCount, *searchKey, searchFields, r.relevanceThresholdConfig)
	}
	// Based on the order field, we apply it on different tables
	if orderBy == flowOrderByRelevance {
		order = mm_db.GenerateFuzzySearchOrderQuery(searchFields, orderDir)
	} else {
		order = fmt.Sprintf("%s %s", orderBy, orderDir)
	}

	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(limit).Offset(offset).Order(order).Find(&models)
	queryCount.Count(&totalCount)

	if result.Error != nil {
		return []flowEntity{}, 0, result.Error
	}
	var entities []flowEntity = []flowEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, totalCount, nil
}

func (r flowRepository) getFlowByID(tx *gorm.DB, flowID uuid.UUID, forUpdate bool) (flowEntity, error) {
	var model *flowModel
	query := tx.Where("id = ?", flowID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return flowEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowRepository) getFlowByCode(tx *gorm.DB, useCaseID uuid.UUID, flowCode string, forUpdate bool) (flowEntity, error) {
	var model *flowModel
	query := tx.Where("use_case_id = ?", useCaseID).Where("code = ?", flowCode)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return flowEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return flowEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r flowRepository) saveFlow(tx *gorm.DB, flow flowEntity, operation mm_db.SaveOperation) (flowEntity, error) {
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

func (r flowRepository) deleteFlow(tx *gorm.DB, flow flowEntity) (flowEntity, error) {
	var model = flowModel(flow)
	err := tx.Delete(model).Error
	if err != nil {
		return flowEntity{}, err
	}
	return flow, nil
}

func (r flowRepository) makeFallbackConsistent(tx *gorm.DB, fallbackFlow flowEntity) error {
	err := tx.Exec(`
		UPDATE mm_flow f
		SET fallback = FALSE
		WHERE f.use_case_id = ?
		AND f.id != ?
	`, fallbackFlow.UseCaseID, fallbackFlow.ID).Error
	if err != nil {
		return err
	}
	return nil

}
