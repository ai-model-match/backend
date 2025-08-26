package useCase

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
	zap.L().Info("Initialize UseCase package...")
	var repository useCaseRepositoryInterface
	var service useCaseServiceInterface
	var router useCaseRouterInterface

	repository = newUseCaseRepository(envs.SearchRelevanceThreshold)
	service = newUseCaseService(dbStorage, pubSubAgent, repository)
	router = newUseCaseRouter(service)
	router.register(routerGroup)
	zap.L().Info("UseCase package initialized")
}
