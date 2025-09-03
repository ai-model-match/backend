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
Each time there is an update on Flow statistics, retrieve the Rollout Strategy related to its Use Case
and evaluate if it needs to be updated (change status or change percentages on Flows).
Evaluation is done on number of requests, number of feedbacks.
This is relevant for Rollout Strategies in WARMUP and ADAPT states to adapt percentages on the fly.
*/
func (s rsEngineService) onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity) error {
	// 1. Retrieve RS based on event.UseCaseID
	// 2. Retrieve all Flows for the Use Case by event.UseCaseID
	// 3. Update each Flow current PCT based on the event.Configuration (use WARMUP or ADAPT configuration based on current state)
	// 4. Valuate if the state of the RS needs to be updated (e.g. WARMUP goal or ADAPT goal is achieved or ESCAPE is needed)
	// 5. Done
	zap.L().Info("onFlowStatisticsUpdate called", zap.String("service", "rs-engine-service"))
	return nil
}

/*
In case a Rollout Strategy is forced to escape, it needs to be handled accordingly adapting
percentages on Flows.
*/
func (s rsEngineService) onRolloutStrategyForcedEscaped(event mm_pubsub.RolloutStrategyEventEntity) error {
	// 1. The event represents the RS
	// 2. Retrieve all Flows for the Use Case by event.UseCaseID
	// 3. Update each Flow current PCT based on the event.Configuration relate to ESCAPE
	// 4. Done
	zap.L().Info("onRolloutStrategyForcedEscaped called", zap.String("service", "rs-engine-service"))
	return nil
}

/*
Each minute retrieve Rollout Strategies that are in WARMUP status and have a timeframe to achieve the warmup goal
This is relevant for Rollout Strategies only in WARMUP state.
*/
func (s rsEngineService) onTimeTick() error {
	// 1. Retrieve all RS in WARMUP status (timing apply only on WARMUP phase). Now for each RS:
	// 2. Retrieve all Flows for the Use Case by RS.UseCaseID
	// 3. Update each Flow current PCT based on the RS.Configuration
	// 4. Valuate if the state of the RS needs to be updated (e.g. WARMUP goal is achieved, not valid for ADAPT state. Or ESCAPE is needed).
	// 5. Done
	zap.L().Info("onTimeTick called", zap.String("service", "rs-engine-service"))
	return nil
}
