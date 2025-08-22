package healthCheck

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type healthCheckServiceInterface interface {
	checkConnection(ctx *gin.Context) (bool, error)
}

type healthCheckService struct {
	storage    *gorm.DB
	repository healthCheckRepositoryInterface
}

func newHealthCheckService(storage *gorm.DB, repository healthCheckRepositoryInterface) healthCheckService {
	return healthCheckService{
		storage:    storage,
		repository: repository,
	}
}

func (s healthCheckService) checkConnection(ctx *gin.Context) (bool, error) {
	return s.repository.checkConnection(s.storage)
}
