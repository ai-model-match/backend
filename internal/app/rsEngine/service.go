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
	onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity, updatedFields []string) error
	onRolloutStrategyChangeState(event mm_pubsub.RolloutStrategyEventEntity) error
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
Each time there is an update on Flow statistics, run the Rollout strategy evaluation on the Use Case
and related Flows tied to this event
*/
func (s rsEngineService) onFlowStatisticsUpdate(event mm_pubsub.FlowStatisticsEventEntity, updatedFields []string) error {
	//
	//	WARMUP Phase to ADAPTIVE Phase
	//
	if mm_utils.SliceContainsAtLeastOneOf([]string{"TotSessionRequests"}, updatedFields) {
		// Retrieve the Rollout Strategy
		rs, err := s.repository.getRolloutStrategyByUseCaseID(s.storage, event.UseCaseID)
		if err != nil {
			return err
		}
		// If the RS does not exist, skip it
		if mm_utils.IsEmpty(rs) {
			return nil
		}
		// If the RS is not in the WARMUP status, skip it
		if rs.RolloutState != mm_pubsub.RolloutStateWarmup {
			return nil
		}
		// If the Warmup configuration is not base on Traffic rules, skip it
		if rs.Configuration.Warmup.IntervalSessReqs == nil {
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
			statistics, err := s.repository.getFlowStatisticsByUseCaseID(tx, rs.UseCaseID)
			if err != nil {
				return err
			}
			for _, stat := range statistics {
				indexedStatistics[stat.FlowID.String()] = stat.TotSessionRequests
				totalCountSessionReqs += stat.TotSessionRequests
			}
			// Retrieve all active Flows for the Use Case
			flows, err := s.repository.getActiveFlowsByUseCaseID(tx, rs.UseCaseID, true)
			if err != nil {
				return err
			}
			// Indicates if all Flows have achieved their goal
			var flowGoalAchieved bool = true
			var totalPctReservedForGoals float64 = 0
			for _, flow := range flows {
				// Per each flow update PCT and check if there is an explicit Warmup goal achieved
				if pctGoal, ok := indexedGoals[flow.ID.String()]; !ok {
					// No explicit Warmup goal, add to the list
					activeFlowsWithoutGoal = append(activeFlowsWithoutGoal, flow)
					continue
				} else {
					// Update the Flow PCT based on statistics
					flow.CurrentServePct = mm_utils.Float64Ptr(calculateNewServePct(*flow.CurrentServePct, pctGoal, totalCountSessionReqs, *rs.Configuration.Warmup.IntervalSessReqs))
					totalPctReservedForGoals += pctGoal
					flow.UpdatedAt = time.Now()
					// Save and check if the goal has been achieved
					s.repository.saveFlow(tx, flow, mm_db.Update)
					if *flow.CurrentServePct != pctGoal {
						flowGoalAchieved = false
					}
				}
			}
			// Now that we updated all Flows with an explicit Warmup goal, proceed with others
			remainingPctPerFlow := (100.0 - totalPctReservedForGoals) / float64(len(activeFlowsWithoutGoal))
			for _, flow := range activeFlowsWithoutGoal {
				flow.CurrentServePct = mm_utils.Float64Ptr(calculateNewServePct(*flow.CurrentServePct, remainingPctPerFlow, totalCountSessionReqs, *rs.Configuration.Warmup.IntervalSessReqs))
				flow.UpdatedAt = time.Now()
				s.repository.saveFlow(tx, flow, mm_db.Update)
				if *flow.CurrentServePct != remainingPctPerFlow {
					flowGoalAchieved = false
				}
			}
			// If all flows have achieved their goal, we can move to the next state
			if flowGoalAchieved {
				rs.RolloutState = mm_pubsub.RolloutStateAdaptive
				rs.UpdatedAt = time.Now()
				s.repository.saveRolloutStrategy(tx, rs, mm_db.Update)
			}
			return nil
		})
		if errTransaction != nil {
			return errTransaction
		}
	}
	//
	//	WARMUP or ADAPTIVE Phase to ESCAPE Phase
	//
	if mm_utils.SliceContainsAtLeastOneOf([]string{"TotFeedback"}, updatedFields) {
		// Retrieve the Rollout Strategy
		rs, err := s.repository.getRolloutStrategyByUseCaseID(s.storage, event.UseCaseID)
		if err != nil {
			return err
		}
		// If the RS does not exist, skip it
		if mm_utils.IsEmpty(rs) {
			return nil
		}
		// If the RS is not in the WARMUP or ADAPTIVE status, skip it
		if rs.RolloutState != mm_pubsub.RolloutStateWarmup && rs.RolloutState != mm_pubsub.RolloutStateAdaptive {
			return nil
		}
		// If the Escape configuration is not defined, skip it
		if rs.Configuration.Escape == nil {
			return nil
		}
		// Start transaction
		errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
			// Representation of Escape rules (FlowID --> Escape Rule)
			indexedRules := map[string]mm_pubsub.RsEscapeRule{}
			for _, rule := range rs.Configuration.Escape.Rules {
				indexedRules[rule.FlowID.String()] = rule
			}
			// Representation of Flow Statistics (FlowID --> Count Session Requests)
			indexedStatistics := map[string]flowStatisticsEntity{}
			statistics, err := s.repository.getFlowStatisticsByUseCaseID(tx, rs.UseCaseID)
			if err != nil {
				return err
			}
			for _, stat := range statistics {
				indexedStatistics[stat.FlowID.String()] = stat
			}
			// Retrieve all active Flows for the Use Case
			flows, err := s.repository.getActiveFlowsByUseCaseID(tx, rs.UseCaseID, true)
			if err != nil {
				return err
			}
			for _, flow := range flows {
				// Per each flow check if there is an explicit Rule for Escape
				if rule, ok := indexedRules[flow.ID.String()]; ok {
					// If yes, so check if there is also a Flow Statistics
					if stat, ok := indexedStatistics[flow.ID.String()]; ok {
						// Check if the Escape rule matches (based on min number of feedback and score)
						if rule.MinFeedback <= stat.TotFeedback && rule.LowerScore >= stat.AvgScore {
							// If yes, move the Rollout Strategy in ESCAPED status
							rs.RolloutState = mm_pubsub.RolloutStateEscaped
							rs.UpdatedAt = time.Now()
							s.repository.saveRolloutStrategy(tx, rs, mm_db.Update)
							// Representation of Escape rules (FlowID --> Escape Rule)
							indexedRollback := map[string]mm_pubsub.RsEscapeRollback{}
							for _, rb := range rule.Rollback {
								indexedRollback[rb.FlowID.String()] = rb
							}
							// Adapt all Flows to the Rollback PCTs
							for _, f := range flows {
								// For each active Flow check if there is a Rollback rule that
								// determine the final Pct, otherwise set to 0
								if rb, ok := indexedRollback[f.ID.String()]; ok {
									f.CurrentServePct = &rb.FinalServePct
								} else {
									f.CurrentServePct = mm_utils.Float64Ptr(0)
								}
								f.UpdatedAt = time.Now()
								s.repository.saveFlow(tx, f, mm_db.Update)
							}
							break
						}
					}
				}
			}
			return nil
		})
		if errTransaction != nil {
			return errTransaction
		}
	}
	return nil
}

/*
In case a Rollout Strategy is forced to escape, run the Rollout strategy evaluation.
*/
func (s rsEngineService) onRolloutStrategyChangeState(event mm_pubsub.RolloutStrategyEventEntity) error {
	rs := rolloutStrategyEntity{
		ID:            event.ID,
		UseCaseID:     event.UseCaseID,
		RolloutState:  event.RolloutState,
		Configuration: event.Configuration,
		UpdatedAt:     event.UpdatedAt,
	}
	// If the RS is not in the FORCED_ESCAPED, skip it
	if rs.RolloutState != mm_pubsub.RolloutStateForcedEscaped {
		return nil
	}
	// If the Escape configuration is not defined, skip it
	if rs.Configuration.Escape == nil {
		return nil
	}
	// Start transaction
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Representation of Escape rules (FlowID --> Escape Rule)
		indexedRules := map[string]mm_pubsub.RsEscapeRule{}
		for _, rule := range rs.Configuration.Escape.Rules {
			indexedRules[rule.FlowID.String()] = rule
		}
		// Retrieve all active Flows for the Use Case
		flows, err := s.repository.getActiveFlowsByUseCaseID(tx, rs.UseCaseID, true)
		if err != nil {
			return err
		}
		for _, flow := range flows {
			// Per each flow check if there is an explicit Rule for Escape
			if rule, ok := indexedRules[flow.ID.String()]; ok {
				// Representation of Escape rules (FlowID --> Escape Rule)
				indexedRollback := map[string]mm_pubsub.RsEscapeRollback{}
				for _, rb := range rule.Rollback {
					indexedRollback[rb.FlowID.String()] = rb
				}
				// Adapt all Flows to the Rollback PCTs
				for _, f := range flows {
					// For each active Flow check if there is a Rollback rule that
					// determine the final Pct, otherwise set to 0
					if rb, ok := indexedRollback[f.ID.String()]; ok {
						f.CurrentServePct = &rb.FinalServePct
					} else {
						f.CurrentServePct = mm_utils.Float64Ptr(0)
					}
					f.UpdatedAt = time.Now()
					s.repository.saveFlow(tx, f, mm_db.Update)
				}
				break
			}
		}
		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
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
