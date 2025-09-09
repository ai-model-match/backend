package rsEngine

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/google/uuid"
)

func calculateNewServePct(currentPct float64, targetPct float64, currentSteps int64, targetSteps int64) float64 {
	// If target PCT of the Flow has been reach, we are fine
	if currentPct == targetPct {
		return targetPct
	}
	// If the total number of session requests on the Use Case is greater or equal to the target,
	// we can force the current PCT to the target PCT
	if currentSteps >= targetSteps {
		return targetPct
	}
	// Otherwise let's calculate the delta to add/sub from the current PCT of the Flow based
	// on how many requests are missing.
	// Adding +1 because when we calculate the delta, the current steps has been already incremented
	// and we are reacting from that, it is not in combination.
	delta := (targetPct - currentPct) / float64(targetSteps-currentSteps+1)
	return mm_utils.RoundTo2Decimals(currentPct + delta)
}

/*
Create a new Event to send for RS Engine update to notify Flows and Rollout Strategy of new changes
based on the different phases of the Engine
*/
func prepareEvent(rs rolloutStrategyEntity, flows []flowEntity) mm_pubsub.PubSubMessage {
	flowEntities := []mm_pubsub.RsEngineFlowEventEntity{}
	for i := range flows {
		flowEntities = append(flowEntities, mm_pubsub.RsEngineFlowEventEntity{
			FlowID:          flows[i].ID,
			CurrentServePct: *flows[i].CurrentServePct,
		})
	}
	eventEntity := &mm_pubsub.RsEngineEventEntity{
		ID:           uuid.New(),
		UseCaseID:    rs.UseCaseID,
		RolloutID:    rs.ID,
		RolloutState: rs.RolloutState,
		Flows:        flowEntities,
	}
	return mm_pubsub.PubSubMessage{
		Message: mm_pubsub.PubSubEvent{
			EventID:            uuid.New(),
			EventTime:          time.Now(),
			EventType:          mm_pubsub.RsEngineUpdatedEvent,
			EventEntity:        eventEntity,
			EventChangedFields: mm_utils.DiffStructs(mm_pubsub.RsEngineEventEntity{}, *eventEntity),
		}}
}
