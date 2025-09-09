package useCaseStep

import (
	"fmt"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type useCaseStepRepositoryInterface interface {
	checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error)
	listUseCaseSteps(tx *gorm.DB, useCaseID uuid.UUID, limit int, offset int, orderBy useCaseStepOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]useCaseStepEntity, int64, error)
	getUseCaseStepByID(tx *gorm.DB, useCaseStepID uuid.UUID, forUpdate bool) (useCaseStepEntity, error)
	getUseCaseStepByCode(tx *gorm.DB, useCaseID uuid.UUID, useCaseStepCode string, forUpdate bool) (useCaseStepEntity, error)
	saveUseCaseStep(tx *gorm.DB, useCaseStep useCaseStepEntity, operation mm_db.SaveOperation) (useCaseStepEntity, error)
	deleteUseCaseStep(tx *gorm.DB, useCaseStep useCaseStepEntity) (useCaseStepEntity, error)
	recalculateUseCaseStepPosition(tx *gorm.DB, useCaseID uuid.UUID) ([]useCaseStepEntity, error)
}

type useCaseStepRepository struct {
	relevanceThresholdConfig float64
}

func newUseCaseStepRepository(relevanceThresholdConfig float64) useCaseStepRepository {
	return useCaseStepRepository{
		relevanceThresholdConfig: relevanceThresholdConfig,
	}
}

func (r useCaseStepRepository) checkUseCaseExists(tx *gorm.DB, useCaseID uuid.UUID) (bool, error) {
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

func (r useCaseStepRepository) listUseCaseSteps(tx *gorm.DB, useCaseID uuid.UUID, limit int, offset int, orderBy useCaseStepOrderBy, orderDir mm_db.OrderDir, searchKey *string, forUpdate bool) ([]useCaseStepEntity, int64, error) {
	var totalCount int64
	var order string

	var models []*useCaseStepModel
	query := tx.Model(useCaseStepModel{}).Where("use_case_id = ?", useCaseID)
	queryCount := tx.Model(useCaseStepModel{}).Where("use_case_id = ?", useCaseID)

	// The ordering of these fields is important for the relevance order
	searchFields := []string{"code", "title", "description"}
	// Add fuzzy search query based on the provided search key and table fields
	if searchKey != nil {
		mm_db.GenerateFuzzySearch(query, *searchKey, searchFields, r.relevanceThresholdConfig)
		mm_db.GenerateFuzzySearch(queryCount, *searchKey, searchFields, r.relevanceThresholdConfig)
	}
	// Based on the order field, we apply it on different tables
	if orderBy == useCaseStepOrderByRelevance {
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
		return []useCaseStepEntity{}, 0, result.Error
	}
	var entities []useCaseStepEntity = []useCaseStepEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, totalCount, nil
}

func (r useCaseStepRepository) getUseCaseStepByID(tx *gorm.DB, useCaseStepID uuid.UUID, forUpdate bool) (useCaseStepEntity, error) {
	var model *useCaseStepModel
	query := tx.Where("id = ?", useCaseStepID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return useCaseStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r useCaseStepRepository) getUseCaseStepByCode(tx *gorm.DB, useCaseID uuid.UUID, useCaseStepCode string, forUpdate bool) (useCaseStepEntity, error) {
	var model *useCaseStepModel
	query := tx.Where("use_case_id = ?", useCaseID).Where("code = ?", useCaseStepCode)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return useCaseStepEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return useCaseStepEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r useCaseStepRepository) saveUseCaseStep(tx *gorm.DB, useCaseStep useCaseStepEntity, operation mm_db.SaveOperation) (useCaseStepEntity, error) {
	var model = useCaseStepModel(useCaseStep)
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
		return useCaseStepEntity{}, err
	}
	return useCaseStep, nil
}

func (r useCaseStepRepository) deleteUseCaseStep(tx *gorm.DB, useCaseStep useCaseStepEntity) (useCaseStepEntity, error) {
	var model = useCaseStepModel(useCaseStep)
	err := tx.Delete(model).Error
	if err != nil {
		return useCaseStepEntity{}, err
	}
	return useCaseStep, nil
}

func (r useCaseStepRepository) recalculateUseCaseStepPosition(tx *gorm.DB, useCaseID uuid.UUID) ([]useCaseStepEntity, error) {
	var models []*useCaseStepModel
	err := tx.Raw(`
		WITH ordered AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position ASC, updated_at ASC) AS new_position
			FROM mm_use_case_step
			WHERE use_case_id = ?
		)
		UPDATE mm_use_case_step s
		SET position = o.new_position, updated_at = NOW()
		FROM ordered o
		WHERE s.id = o.id
		AND s.position IS DISTINCT FROM o.new_position
		RETURNING s.*;
	`, useCaseID).Scan(&models).Error
	if err != nil {
		return []useCaseStepEntity{}, err
	}
	// Return only updated entities
	var entities []useCaseStepEntity = []useCaseStepEntity{}
	for _, model := range models {
		entity := model.toEntity()
		entities = append(entities, entity)
	}
	return entities, nil
}
