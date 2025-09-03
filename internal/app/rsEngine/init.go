package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs and PubSub consumers.
*/
func Init(envs *mm_env.Envs, dbStorage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, cron *mm_scheduler.Scheduler) {
	zap.L().Info("Initialize RsEngine package...")
	var repository rsEngineRepositoryInterface
	var service rsEngineServiceInterface
	var scheduler rsEngineSchedulerInterface
	var consumer rsEngineConsumerInterface

	repository = newRsEngineRepository()
	service = newRsEngineService(dbStorage, pubSubAgent, repository)
	scheduler = newRsEngineScheduler(dbStorage, cron, service)
	consumer = newRsEngineConsumer(pubSubAgent, service)
	scheduler.init()
	consumer.subscribe()
	zap.L().Info("RsEngine package initialized")
}
