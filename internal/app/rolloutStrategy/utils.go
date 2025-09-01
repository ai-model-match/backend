package rolloutStrategy

import (
	"slices"

	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
)

/*
Check if possible to move the Flow State to the next selected one based on the state machine defined
*/
func checkStateFlow(currentState mm_pubsub.RolloutState, nextState mm_pubsub.RolloutState) bool {
	if nextStates, ok := allowedTransitions[currentState]; ok {
		if slices.Contains(nextStates, nextState) {
			return true
		}
	}
	return false
}
