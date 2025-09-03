package rsEngine

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
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
	// 2. Ensure the RS Warmup is traffic base, otherwise return
	// 3. Retrieve all Flows for the Use Case by event.UseCaseID
	// 4. Update each Flow current PCT based on the event.Configuration.Warmup
	// 5. Valuate if the state of the RS needs to be updated (e.g. WARMUP to ADAPTIVE or WARMUP to ESCAPE)
	// 6. Done
	rs, err := s.repository.getRolloutStrategyByUseCaseIDAndStatus(s.storage, event.UseCaseID, mm_pubsub.RolloutStateWarmup)
	if err != nil {
		return err
	}
	if mm_utils.IsEmpty(rs) {
		return errRolloutStrategyNotFound
	}
	// Proceed only for RS configuration that require traffic based Warmup
	if rs.Configuration.Warmup.IntervalSessReqs == nil {
		zap.L().Info("onFlowStatisticsUpdate skipped, not traffic based Warmup", zap.String("service", "rs-engine-service"), zap.String("useCaseID", event.UseCaseID.String()))
		return nil
	}
	// Start transaction
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		indexedFlow := map[string]flowEntity{}
		flows, err := s.repository.getFlowsByUseCaseID(s.storage, rs.UseCaseID, true)
		if err != nil {
			return err
		}
		for _, flow := range flows {
			indexedFlow[flow.ID.String()] = flow
		}
		indexedStatistics := map[string]flowStatisticsEntity{}
		statistics, err := s.repository.getFlowStatisticsByUseCaseID(s.storage, rs.UseCaseID)
		if err != nil {
			return err
		}
		var totReq int64 = 0
		for _, stat := range statistics {
			indexedStatistics[stat.FlowID.String()] = stat
			totReq += stat.TotSessionRequests
		}
		var goalAchieved bool = true
		for _, goal := range rs.Configuration.Warmup.Goals {
			if flow, ok := indexedFlow[goal.FlowID.String()]; ok {
				if goal.FinalServePct == flow.CurrentServePct {
					continue
				}
				delta := (goal.FinalServePct - flow.CurrentServePct) / float64(*rs.Configuration.Warmup.IntervalSessReqs-(totReq-1))
				flow.CurrentServePct = mm_utils.RoundTo2Decimals(flow.CurrentServePct + delta)
				if flow.CurrentServePct != goal.FinalServePct {
					goalAchieved = false
				}
				flow.UpdatedAt = time.Now()
				s.repository.saveFlow(tx, flow, mm_db.Update)
			}
		}
		if goalAchieved {
			zap.L().Info("onFlowStatisticsUpdate, all goals achieved, moving to ADAPTIVE", zap.String("service", "rs-engine-service"), zap.String("useCaseID", event.UseCaseID.String()))
			rs.RolloutState = mm_pubsub.RolloutStateAdaptive
			rs.UpdatedAt = time.Now()
			s.repository.saveRolloutStrategy(tx, rs, mm_db.Update)
		}

		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
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

func (s rsEngineService) getRolloutStrategyByUseCaseIDAndStatus(useCaseID string, status mm_pubsub.RolloutState) error {
	return nil
}

func (s rsEngineService) getRolloutStrategies() error {
	return nil
}
