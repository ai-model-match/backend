package flowStepStatistics

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
	zap.L().Info("Initialize FlowStepStatistics package...")
	var repository flowStepStatisticsRepositoryInterface
	var service flowStepStatisticsServiceInterface
	var router flowStepStatisticsRouterInterface
	var consumer flowStepStatisticsConsumerInterface

	repository = newFlowStepStatisticsRepository()
	service = newFlowStepStatisticsService(dbStorage, pubSubAgent, repository)
	router = newFlowStepStatisticsRouter(service)
	consumer = newFlowStepStatisticsConsumer(pubSubAgent, service)
	consumer.subscribe()
	router.register(routerGroup)
	zap.L().Info("FlowStepStatistics package initialized")
}
