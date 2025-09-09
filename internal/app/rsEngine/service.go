package rsEngine

import (
	"math"
	"slices"
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
	// If the RS is FORCED_ESCAPED
	if rs.RolloutState == mm_pubsub.RolloutStateForcedEscaped {
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
	}
	// If the RS is FORCED_COMPLETED
	if rs.RolloutState == mm_pubsub.RolloutStateForcedCompleted {
		// If the Escape configuration is not defined, skip it
		if rs.Configuration.StateConfigurations.CompletedFlowID == nil {
			return nil
		}
		// Start transaction
		errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
			forcedFlowID := *rs.Configuration.StateConfigurations.CompletedFlowID

			// Retrieve all active Flows for the Use Case
			flows, err := s.repository.getActiveFlowsByUseCaseID(tx, rs.UseCaseID, true)
			if err != nil {
				return err
			}
			// Per each flow, if equal to the forced Flow, put to 100%, otherwise to 0%
			for _, flow := range flows {
				if flow.ID == forcedFlowID {
					flow.CurrentServePct = mm_utils.Float64Ptr(100)
				} else {
					flow.CurrentServePct = mm_utils.Float64Ptr(0)
				}
				flow.UpdatedAt = time.Now()
				s.repository.saveFlow(tx, flow, mm_db.Update)

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
Each minute check which Rollout Strategies need to be evaluated and proceed.
*/
func (s rsEngineService) onTimeTick() error {
	rolloutStrategies, err := s.repository.getActiveRolloutStrategiesInState(s.storage, []mm_pubsub.RolloutState{mm_pubsub.RolloutStateWarmup, mm_pubsub.RolloutStateAdaptive})
	if err != nil {
		return err
	}
	// For each Rollout Strategy, run their warmup or adaptive phase
	for _, rs := range rolloutStrategies {
		go func() {
			if err := s.tickOnRolloutStrategy(rs); err != nil {
				zap.L().Error("Something went wrong during RS Engine execution", zap.String("Use Case ID", rs.UseCaseID.String()), zap.Error(err), zap.String("service", "rs-engine-service"))
			}
		}()
	}
	return nil
}

func (s rsEngineService) tickOnRolloutStrategy(rs rolloutStrategyEntity) error {
	//
	//	WARMUP Phase
	//
	if rs.RolloutState == mm_pubsub.RolloutStateWarmup {
		// If the Warmup configuration is not base on Time rules, skip it
		if rs.Configuration.Warmup.IntervalMins == nil {
			return nil
		}
		// List of Flows without an explicit Warmup goal
		activeFlowsWithoutGoal := []flowEntity{}
		// Representation of Warmup goal (FlowID --> Pct Goal)
		indexedGoals := map[string]float64{}
		for _, goal := range rs.Configuration.Warmup.Goals {
			indexedGoals[goal.FlowID.String()] = goal.FinalServePct
		}
		// Start transaction
		errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
			// Retrieve all active Flows for the Use Case
			flows, err := s.repository.getActiveFlowsByUseCaseID(tx, rs.UseCaseID, true)
			if err != nil {
				return err
			}
			// From now, check how many minutes are missing to reach the target (start of WARMUP + Interval Mins)
			missingMinutes := int64(math.Round(time.Until((rs.UpdatedAt.Add(time.Duration(*rs.Configuration.Warmup.IntervalMins * int64(time.Minute))))).Minutes()))
			// Specific case, should never happen, force to 0, so move automatically to ADAPTIVE
			if missingMinutes < 0 {
				missingMinutes = 0
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
					flow.CurrentServePct = mm_utils.Float64Ptr(calculateNewServePct(*flow.CurrentServePct, pctGoal, 0, missingMinutes))
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
				flow.CurrentServePct = mm_utils.Float64Ptr(calculateNewServePct(*flow.CurrentServePct, remainingPctPerFlow, 0, missingMinutes))
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
		// Avoid Adaptive phase is executed immediately once this one is completed
		return nil
	}

	//
	//	ADAPTIVE Phase
	//
	if rs.RolloutState == mm_pubsub.RolloutStateAdaptive {
		elaspedMinutes := int64(math.Round(time.Since(rs.UpdatedAt).Minutes()))
		// If elasped time is not a multiple of the interval, skip it
		if elaspedMinutes%rs.Configuration.Adaptive.IntervalMins != 0 {
			return nil
		}
		// Start transaction
		errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
			// Representation of Flow Statistics (FlowID --> Stats)
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
			// Check for each flow if it has at least the number of needed feeedback
			// to start the adaptive phase
			isAdaptiveReady := true
			for _, flow := range flows {
				if stat, ok := indexedStatistics[flow.ID.String()]; ok {
					if stat.TotFeedback < rs.Configuration.Adaptive.MinFeedback {
						isAdaptiveReady = false
					}
				}
			}
			// Check if we can proceed with Adaptive phase
			if !isAdaptiveReady {
				return nil
			}

			// Find highest score and best Flow Indexes
			bestScore := 0.0
			bestFlowIndexes := []int{}
			for i := range flows {
				if stat, ok := indexedStatistics[flows[i].ID.String()]; ok {
					if stat.AvgScore > bestScore {
						bestScore = stat.AvgScore
						bestFlowIndexes = []int{i}
					} else if stat.AvgScore == bestScore {
						bestFlowIndexes = append(bestFlowIndexes, i)
					}
				}
			}
			// Check how much traffic is provided by worst Flows and find their Indexes
			servedByWorst := 0.0
			sumAvgScoreByWorst := 0.0
			worstFlowIndexes := []int{}
			for i := range flows {
				if !slices.Contains(bestFlowIndexes, i) {
					worstFlowIndexes = append(worstFlowIndexes, i)
					servedByWorst += *flows[i].CurrentServePct
					if stat, ok := indexedStatistics[flows[i].ID.String()]; ok {
						sumAvgScoreByWorst += stat.AvgScore
					}
				}
			}
			sumAvgScoreByWorst = mm_utils.RoundTo2Decimals(sumAvgScoreByWorst)
			// Calculate the maximum increment we can assign to best Flows by decrementing worst Flows
			// keeping as limit the provided configuration
			totalPossibleIncrement := rs.Configuration.Adaptive.MaxStepPct
			if servedByWorst < totalPossibleIncrement {
				totalPossibleIncrement = servedByWorst
			}
			// Calculate the increment per Flow (in case there are multiple best flows, split the increment between them)
			incrementPerBestFlow := totalPossibleIncrement / float64(len(bestFlowIndexes))
			// Increase traffic for best flows, without passing 100%. Track if one Flow reached 100%
			flowReachedMaxPct := false
			for _, bestFlowIndex := range bestFlowIndexes {
				newPct := *flows[bestFlowIndex].CurrentServePct + incrementPerBestFlow
				flows[bestFlowIndex].CurrentServePct = mm_utils.RoundTo2DecimalsPtr(&newPct)
				if *flows[bestFlowIndex].CurrentServePct >= 100 {
					flowReachedMaxPct = true
					*flows[bestFlowIndex].CurrentServePct = 100
				}
			}
			// If one Flow reached 100%, we need to put all others to 0%
			if flowReachedMaxPct {
				for i := range flows {
					if *flows[i].CurrentServePct != 100.0 {
						flows[i].CurrentServePct = mm_utils.Float64Ptr(0)
					}
				}
			} else {
				// Otherwise, for each worst Flow, calculate how much we need to decrement it based on the distance between its Score and Best Score
				totDecremented := 0.0
				for _, i := range worstFlowIndexes {
					if stat, ok := indexedStatistics[flows[i].ID.String()]; ok {
						toDecrement := ((sumAvgScoreByWorst - stat.AvgScore) / (sumAvgScoreByWorst * float64(len(worstFlowIndexes)-1)) * totalPossibleIncrement)
						newPct := *flows[i].CurrentServePct - toDecrement

						if newPct < 0.0 {
							totDecremented += *flows[i].CurrentServePct
							flows[i].CurrentServePct = mm_utils.Float64Ptr(0)
						} else {
							totDecremented += toDecrement
							flows[i].CurrentServePct = mm_utils.RoundTo2DecimalsPtr(&newPct)
						}
					}
				}
				if totalPossibleIncrement-totDecremented > 0 {
					missingDecrement := totalPossibleIncrement - totDecremented
					// If there is additional PCT to decrement, proceed in order, by sorting all Flows based on their Scores
					slices.SortFunc(flows, func(a flowEntity, b flowEntity) int {
						if statA, okA := indexedStatistics[a.ID.String()]; okA {
							if statB, okB := indexedStatistics[b.ID.String()]; okB {
								return int((statA.AvgScore - statB.AvgScore) * 100)
							}
						}
						return 0
					})
					// Now, considering only Worst Flows, start reducing the traffic frome the worst Flow to better one
					for i := range flows {
						if missingDecrement == 0 {
							break
						}
						if stat, ok := indexedStatistics[flows[i].ID.String()]; ok {
							// Consider only worst Flows
							if stat.AvgScore != bestScore {
								if *flows[i].CurrentServePct >= missingDecrement {
									newPct := *flows[i].CurrentServePct - missingDecrement
									flows[i].CurrentServePct = mm_utils.RoundTo2DecimalsPtr(&newPct)
									missingDecrement = 0
								} else {
									missingDecrement -= *flows[i].CurrentServePct
									flows[i].CurrentServePct = mm_utils.Float64Ptr(0)
								}
							}
						}
					}
				}
			}
			// Save all Flows
			for i := range flows {
				flows[i].UpdatedAt = time.Now()
				s.repository.saveFlow(tx, flows[i], mm_db.Update)
			}

			// If the Adaptive phase achieved its goal, move to COMPLETED
			if flowReachedMaxPct {
				rs.RolloutState = mm_pubsub.RolloutStateCompleted
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
	return nil
}
