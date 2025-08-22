package auth

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type authRepositoryInterface interface {
	getAuthSessionEntityByRefreshToken(tx *gorm.DB, refreshToken string, forUpdate bool) (authSessionEntity, error)
	saveAuthSessionEntity(tx *gorm.DB, entity authSessionEntity) (authSessionEntity, error)
	deleteAuthSessionEntity(tx *gorm.DB, entity authSessionEntity) error
}

type authRepository struct {
}

func newAuthRepository() authRepository {
	return authRepository{}
}

func (r authRepository) getAuthSessionEntityByRefreshToken(tx *gorm.DB, refreshToken string, forUpdate bool) (authSessionEntity, error) {
	var model *authSessionModel
	query := tx.Where("refresh_token = ?", refreshToken)
	query.Where("expires_at > now()")
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return authSessionEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return authSessionEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r authRepository) saveAuthSessionEntity(tx *gorm.DB, entity authSessionEntity) (authSessionEntity, error) {
	var model = authSessionModel(entity)
	err := tx.Save(model).Error
	if err != nil {
		return authSessionEntity{}, err
	}
	return entity, nil
}

func (r authRepository) deleteAuthSessionEntity(tx *gorm.DB, entity authSessionEntity) error {
	var model = authSessionModel(entity)
	if err := tx.Delete(model).Error; err != nil {
		return err
	}
	return nil
}
