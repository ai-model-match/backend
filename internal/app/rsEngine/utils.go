package rsEngine

import "github.com/ai-model-match/backend/internal/pkg/mm_utils"

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
