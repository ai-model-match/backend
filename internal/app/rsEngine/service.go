package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type rsEngineServiceInterface interface {
	onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity) error
	onTimeTick() error
}

type rsEngineService struct {
	storage     *gorm.DB
	pubSubAgent *mm_pubsub.PubSubAgent
	repository  rsEngineRepositoryInterface
}

func newRsEngineService(storage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, repository rsEngineRepositoryInterface) rsEngineService {
	return rsEngineService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}
func (s rsEngineService) onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity) error {
	zap.L().Info("onFlowStatisticsUpdate called", zap.String("service", "rs-engine-service"))
	return nil
}
func (s rsEngineService) onTimeTick() error {
	zap.L().Info("onTimeTick called", zap.String("service", "rs-engine-service"))
	return nil
}
