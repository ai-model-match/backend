package flowStepStatistics

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type flowStepStatisticsRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type flowStepStatisticsRouter struct {
	service flowStepStatisticsServiceInterface
}

func newFlowStepStatisticsRouter(service flowStepStatisticsServiceInterface) flowStepStatisticsRouter {
	return flowStepStatisticsRouter{
		service: service,
	}
}

// Implementation
func (r flowStepStatisticsRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/flow-steps/:flowStepId/flow-step-statistics",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getFlowStepStatisticsInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getFlowStepStatisticsByID(ctx, request)
			if err == errFlowStepStatisticsNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-step-statistics-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
