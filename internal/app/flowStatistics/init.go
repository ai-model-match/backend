package flowStatistics

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
	zap.L().Info("Initialize FlowStatistics package...")
	var repository flowStatisticsRepositoryInterface
	var service flowStatisticsServiceInterface
	var router flowStatisticsRouterInterface
	var consumer flowStatisticsConsumerInterface

	repository = newFlowStatisticsRepository()
	service = newFlowStatisticsService(dbStorage, pubSubAgent, repository)
	router = newFlowStatisticsRouter(service)
	consumer = newFlowStatisticsConsumer(pubSubAgent, service)
	consumer.subscribe()
	router.register(routerGroup)
	zap.L().Info("FlowStatistics package initialized")
}
