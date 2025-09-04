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

	// Retrieve the Rollout Strategy based on Use Case ID and the WARMUP status (the only to be considered for traffic based)
	rs, err := s.repository.getRolloutStrategyByUseCaseIDAndStatus(s.storage, event.UseCaseID, mm_pubsub.RolloutStateWarmup)
	if err != nil {
		return err
	}
	// The RS is on a different state (no WARMUP)
	if mm_utils.IsEmpty(rs) {
		return nil
	}
	// Proceed only for RS configuration that require traffic based Warmup
	if rs.Configuration.Warmup.IntervalSessReqs == nil {
		zap.L().Info("onFlowStatisticsUpdate skipped, not traffic based Warmup", zap.String("useCaseID", event.UseCaseID.String()), zap.String("service", "rs-engine-service"))
		return nil
	}
	// Start transaction
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {

		// List of Flows without an explicit Warmup goal
		activeFlowsWithoutGoal := []flowEntity{}

		// Representation of Warmup goal (FlowID --> Pct Goal)
		indexedGoals := map[string]float64{}
		for _, goal := range rs.Configuration.Warmup.Goals {
			indexedGoals[goal.FlowID.String()] = goal.FinalServePct
		}

		// Representation of Flow Statistics (FlowID --> Count Session Requests)
		indexedStatistics := map[string]int64{}
		// Total Count of Session Requests across all existing Flows
		// Note: inactive Flows are included as well because they can be disabled in the middle, but the toal requests remains.
		var totalCountSessionReqs int64 = 0
		statistics, err := s.repository.getFlowStatisticsByUseCaseID(s.storage, rs.UseCaseID)
		if err != nil {
			return err
		}
		for _, stat := range statistics {
			indexedStatistics[stat.FlowID.String()] = stat.TotSessionRequests
			totalCountSessionReqs += stat.TotSessionRequests
		}

		// Retrieve all active Flows for the Use Case
		flows, err := s.repository.getActiveFlowsByUseCaseID(s.storage, rs.UseCaseID, true)
		if err != nil {
			return err
		}

		// Indicates if all Flows have achieved their goal
		var flowGoalAchieved bool = true
		var totalPctReservedForGoals float64 = 0
		for _, flow := range flows {
			// Per each flow check if there is an explicit Warmup goal
			if pctGoal, ok := indexedGoals[flow.ID.String()]; !ok {
				// No explicit Warmup goal, add to the list
				activeFlowsWithoutGoal = append(activeFlowsWithoutGoal, flow)
				continue
			} else {
				// Update the Flow PCT based on statistics
				flow.CurrentServePct = calculateNewServePct(flow.CurrentServePct, pctGoal, totalCountSessionReqs, *rs.Configuration.Warmup.IntervalSessReqs)
				totalPctReservedForGoals += pctGoal
				flow.UpdatedAt = time.Now()
				s.repository.saveFlow(tx, flow, mm_db.Update)
				if flow.CurrentServePct != pctGoal {
					flowGoalAchieved = false
				}
			}
		}

		// Now that we updated all Flows with an explicit Warmup goal, proceed with others
		remainingPctPerFlow := (100.0 - totalPctReservedForGoals) / float64(len(activeFlowsWithoutGoal))
		for _, flow := range activeFlowsWithoutGoal {
			flow.CurrentServePct = calculateNewServePct(flow.CurrentServePct, remainingPctPerFlow, totalCountSessionReqs, *rs.Configuration.Warmup.IntervalSessReqs)
			flow.UpdatedAt = time.Now()
			s.repository.saveFlow(tx, flow, mm_db.Update)
			if flow.CurrentServePct != remainingPctPerFlow {
				flowGoalAchieved = false
			}
		}

		// If all flows have achieved their goal, we can move to the next state
		if flowGoalAchieved {
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
