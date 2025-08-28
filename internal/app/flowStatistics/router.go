package flowStatistics

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type flowStatisticsRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type flowStatisticsRouter struct {
	service flowStatisticsServiceInterface
}

func newFlowStatisticsRouter(service flowStatisticsServiceInterface) flowStatisticsRouter {
	return flowStatisticsRouter{
		service: service,
	}
}

// Implementation
func (r flowStatisticsRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/flows/:flowId/flow-statistics",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getFlowStatisticsInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getFlowStatisticsByID(ctx, request)
			if err == errFlowStatisticsNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-statistics-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
