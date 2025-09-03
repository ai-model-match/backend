package rsEngine

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type rsEngineServiceInterface interface {
	onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity) error
	onRolloutStrategyForcedEscaped(event mm_pubsub.RolloutStrategyEventEntity) error
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

/*
Each time there is an update on Flow statistics, run the Rollout strategy evaluation.
*/
func (s rsEngineService) onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity) error {
	// 1. Retrieve RS based on event.UseCaseID and WARMUP status
	// 2. Retrieve all Flows for the Use Case by event.UseCaseID
	// 3. Update each Flow current PCT based on the event.Configuration.Warmup
	// 4. Valuate if the state of the RS needs to be updated (e.g. WARMUP to ADAPTIVE or WARMUP to ESCAPE)
	// 5. Done
	zap.L().Info("onFlowStatisticsUpdate called", zap.String("service", "rs-engine-service"))
	return nil
}

/*
In case a Rollout Strategy is forced to escape, run the Rollout strategy evaluation.
*/
func (s rsEngineService) onRolloutStrategyForcedEscaped(event mm_pubsub.RolloutStrategyEventEntity) error {
	// 1. The event represents the RS
	// 2. Retrieve all Flows for the Use Case by event.UseCaseID
	// 3. Update each Flow current PCT based on the event.Configuration.Escape
	// 4. Done
	zap.L().Info("onRolloutStrategyForcedEscaped called", zap.String("service", "rs-engine-service"))
	return nil
}

/*
Each minute check which Rollout Strategies need to be evaluated and proceed.
*/
func (s rsEngineService) onTimeTick() error {
	// 1. Retrieve all RS in WARMUP having a RS.Configuration.Warmup based on timing or in ADAPTIVE status. Now for each RS:
	// 2. Retrieve all Flows for the Use Case by RS.UseCaseID
	// 3. Update each Flow current PCT based on the RS.Configuration.Warmup or RS.Configuration.Adaptive based on the state
	// 4. Valuate if the state of the RS needs to be updated (e.g. WARMUP to ADAPTIVE or WARMUP to ESCAPE or ADAPTIVE to ESCAPE or ADAPTIVE to COMPLETED).
	// 5. Done
	zap.L().Info("onTimeTick called", zap.String("service", "rs-engine-service"))
	return nil
}

func (s rsEngineService) getRolloutStrategyByUseCaseIDAndStatus(useCaseID string, status RolloutState) error {
	return nil
}

func (s rsEngineService) getRolloutStrategies() error {
	return nil
}
