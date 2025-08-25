package flow

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
	zap.L().Info("Initialize Flow package...")
	var repository flowRepositoryInterface
	var service flowServiceInterface
	var router flowRouterInterface

	repository = newFlowRepository(envs.SearchRelevanceThreshold)
	service = newFlowService(dbStorage, pubSubAgent, repository)
	router = newFlowRouter(service)
	router.register(routerGroup)
	zap.L().Info("Flow package initialized")
}
