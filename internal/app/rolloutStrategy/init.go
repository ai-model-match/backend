package rolloutStrategy

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
	zap.L().Info("Initialize RolloutStrategy package...")
	var repository rolloutStrategyRepositoryInterface
	var service rolloutStrategyServiceInterface
	var router rolloutStrategyRouterInterface
	var consumer rolloutStrategyConsumerInterface

	repository = newRolloutStrategyRepository()
	service = newRolloutStrategyService(dbStorage, pubSubAgent, repository)
	router = newRolloutStrategyRouter(service)
	consumer = newRolloutStrategyConsumer(pubSubAgent, service)
	consumer.subscribe()
	router.register(routerGroup)
	zap.L().Info("RolloutStrategy package initialized")
}
