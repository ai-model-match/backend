package healthCheck

import (
	"github.com/ai-model-match/backend/internal/pkg/mmenv"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs
*/
func Init(envs *mmenv.Envs, dbStorage *gorm.DB, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize Health Check package...")
	var repository healthCheckRepositoryInterface
	var service healthCheckServiceInterface
	var router healthCheckRouterInterface

	repository = newHealthCheckRepository()
	service = newHealthCheckService(dbStorage, repository)
	router = newHealthCheckRouter(service)
	router.register(routerGroup)
	zap.L().Info("Health Check package initialized")
}
