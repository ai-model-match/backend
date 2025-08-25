package useCaseStep

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs and PubSub consumers.
*/
func Init(envs *mm_env.Envs, dbStorage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize UseCaseStep package...")
	var repository useCaseStepRepositoryInterface
	var service useCaseStepServiceInterface
	var router useCaseStepRouterInterface

	repository = newUseCaseStepRepository(envs.SearchRelevanceThreshold)
	service = newUseCaseStepService(dbStorage, pubSubAgent, repository)
	router = newUseCaseStepRouter(service)
	router.register(routerGroup)
	zap.L().Info("UseCaseStep package initialized")
}
