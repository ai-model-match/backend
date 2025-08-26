package flowStep

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
	zap.L().Info("Initialize FlowStep package...")
	var repository flowStepRepositoryInterface
	var service flowStepServiceInterface
	var router flowStepRouterInterface
	var consumer flowStepConsumerInterface

	repository = newFlowStepRepository()
	service = newFlowStepService(dbStorage, pubSubAgent, repository)
	router = newFlowStepRouter(service)
	consumer = newFlowStepConsumer(pubSubAgent, service)
	consumer.subscribe()
	router.register(routerGroup)
	zap.L().Info("FlowStep package initialized")
}
