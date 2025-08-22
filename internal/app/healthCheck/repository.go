package healthCheck

import (
	"gorm.io/gorm"
)

type healthCheckRepositoryInterface interface {
	checkConnection(tx *gorm.DB) (bool, error)
}

type healthCheckRepository struct {
}

func newHealthCheckRepository() healthCheckRepository {
	return healthCheckRepository{}
}

func (r healthCheckRepository) checkConnection(tx *gorm.DB) (bool, error) {
	sqlDB, err := tx.DB()
	if err != nil {
		return false, err
	}
	if err := sqlDB.Ping(); err != nil {
		return false, err
	}

	return true, nil
}
