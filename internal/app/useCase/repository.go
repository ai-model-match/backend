package useCase

import (
	"fmt"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type useCaseRepositoryInterface interface {
	listUseCases(tx *gorm.DB, limit int, offset int, orderBy useCaseOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]useCaseEntity, int64, error)
	getUseCaseByID(tx *gorm.DB, useCaseID uuid.UUID, forUpdate bool) (useCaseEntity, error)
	getUseCaseByCode(tx *gorm.DB, useCaseCode string, forUpdate bool) (useCaseEntity, error)
	saveUseCase(tx *gorm.DB, useCase useCaseEntity, operation mm_db.SaveOperation) (useCaseEntity, error)
	deleteUseCase(tx *gorm.DB, useCase useCaseEntity) (useCaseEntity, error)
	checkFallbackFlowExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error)
}

type useCaseRepository struct {
	relevanceThresholdConfig float64
}

func newUseCaseRepository(relevanceThresholdConfig float64) useCaseRepository {
	return useCaseRepository{
		relevanceThresholdConfig: relevanceThresholdConfig,
	}
}

func (r useCaseRepository) listUseCases(tx *gorm.DB, limit int, offset int, orderBy useCaseOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]useCaseEntity, int64, error) {
	var totalCount int64
	var order string

	var models []*useCaseModel
	query := tx.Model(useCaseModel{})
	queryCount := tx.Model(useCaseModel{})

	// The ordering of these fields is important for the relevance order
	searchFields := []string{"code", "title", "description"}
	// Add fuzzy search query based on the provided search key and table fields
	if searchKey != nil {
		mm_db.GenerateFuzzySearch(query, *searchKey, searchFields, r.relevanceThresholdConfig)
		mm_db.GenerateFuzzySearch(queryCount, *searchKey, searchFields, r.relevanceThresholdConfig)
	}
	// Based on the order field, we apply it on different tables
	if orderBy == useCaseOrderByRelevance {
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
		return []useCaseEntity{}, 0, result.Error
	}
	var entities []useCaseEntity = []useCaseEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, totalCount, nil
}

func (r useCaseRepository) getUseCaseByID(tx *gorm.DB, useCaseID uuid.UUID, forUpdate bool) (useCaseEntity, error) {
	var model *useCaseModel
	query := tx.Where("id = ?", useCaseID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return useCaseEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r useCaseRepository) getUseCaseByCode(tx *gorm.DB, useCaseCode string, forUpdate bool) (useCaseEntity, error) {
	var model *useCaseModel
	query := tx.Where("code = ?", useCaseCode)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return useCaseEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r useCaseRepository) saveUseCase(tx *gorm.DB, useCase useCaseEntity, operation mm_db.SaveOperation) (useCaseEntity, error) {
	var model = useCaseModel(useCase)
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
		return useCaseEntity{}, err
	}
	return useCase, nil
}

func (r useCaseRepository) deleteUseCase(tx *gorm.DB, useCase useCaseEntity) (useCaseEntity, error) {
	var model = useCaseModel(useCase)
	err := tx.Delete(model).Error
	if err != nil {
		return useCaseEntity{}, err
	}
	return useCase, nil
}

func (r useCaseRepository) checkFallbackFlowExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error) {
	var model *flowModel
	query := tx.Where("use_case_id = ?", useCaseID).Where("fallback is true")
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 || mm_utils.IsEmpty(model) {
		return false, nil
	}
	return model.Fallback, nil
}
