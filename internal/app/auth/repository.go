package auth

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type authRepositoryInterface interface {
	getAuthEntityByRefreshToken(tx *gorm.DB, refreshToken string, forUpdate bool) (authEntity, error)
	save(tx *gorm.DB, entity authEntity) (authEntity, error)
	delete(tx *gorm.DB, entity authEntity) error
}

type authRepository struct {
}

func newAuthRepository() authRepository {
	return authRepository{}
}

func (r authRepository) getAuthEntityByRefreshToken(tx *gorm.DB, refreshToken string, forUpdate bool) (authEntity, error) {
	var model *authModel
	query := tx.Where("refresh_token = ?", refreshToken)
	query.Where("expires_at > now()")
	if forUpdate {
		query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return authEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return authEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r authRepository) save(tx *gorm.DB, entity authEntity) (authEntity, error) {
	var model = authModel(entity)
	err := tx.Save(model).Error
	if err != nil {
		return authEntity{}, err
	}
	return entity, nil
}

func (r authRepository) delete(tx *gorm.DB, entity authEntity) error {
	var model = authModel(entity)
	if err := tx.Delete(model).Error; err != nil {
		return err
	}
	return nil
}
