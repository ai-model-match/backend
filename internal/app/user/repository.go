package user

import (
	"fmt"

	"github.com/ai-model-match/backend/internal/pkg/mmdb"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepositoryInterface interface {
	listUsers(tx *gorm.DB, limit int, offset int, orderBy userOrderBy, orderDir mmdb.OrderDir, searchKey *string, includeDeleted bool, forUpdate bool) ([]userEntity, int64, error)
	getUserByID(tx *gorm.DB, userID uuid.UUID, forUpdate bool) (userEntity, error)
	saveUser(tx *gorm.DB, user userEntity) (userEntity, error)
}

type userRepository struct {
	relevanceThresholdConfig float64
}

func newUserRepository(relevanceThresholdConfig float64) userRepository {
	return userRepository{
		relevanceThresholdConfig: relevanceThresholdConfig,
	}
}

func (r userRepository) listUsers(tx *gorm.DB, limit int, offset int, orderBy userOrderBy, orderDir mmdb.OrderDir, searchKey *string, includeDeleted bool, forUpdate bool) ([]userEntity, int64, error) {
	var totalCount int64
	var order string

	var models []*userModel
	query := tx.Model(userModel{})
	queryCount := tx.Model(userModel{})

	// The ordering of these fields is important for the relevance order
	searchFields := []string{"email", "lastname", "firstname"}
	// Add fuzzy search query based on the provided search key and table fields
	if searchKey != nil {
		mmdb.GenerateFuzzySearch(query, *searchKey, searchFields, r.relevanceThresholdConfig)
		mmdb.GenerateFuzzySearch(queryCount, *searchKey, searchFields, r.relevanceThresholdConfig)
	}
	// Based on the order field, we apply it on different tables
	if orderBy == userOrderByRelevance {
		order = mmdb.GenerateFuzzySearchOrderQuery(searchFields, orderDir)
	} else {
		order = fmt.Sprintf("%s %s", orderBy, orderDir)
	}

	if !includeDeleted {
		query.Where("deleted_at IS NULL")
		queryCount.Where("deleted_at IS NULL")
	}

	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(limit).Offset(offset).Order(order).Find(&models)
	queryCount.Count(&totalCount)

	if result.Error != nil {
		return []userEntity{}, 0, result.Error
	}
	var entities []userEntity = []userEntity{}
	for _, model := range models {
		fscInput := model.toEntity()
		entities = append(entities, fscInput)
	}
	return entities, totalCount, nil
}

func (r userRepository) getUserByID(tx *gorm.DB, userID uuid.UUID, forUpdate bool) (userEntity, error) {
	var model *userModel
	query := tx.Where("id = ?", userID)
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return userEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return userEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r userRepository) saveUser(tx *gorm.DB, user userEntity) (userEntity, error) {
	var model = userModel(user)
	err := tx.Save(model).Error
	if err != nil {
		return userEntity{}, err
	}
	return user, nil
}
