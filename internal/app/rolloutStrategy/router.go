package rolloutStrategy

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type rolloutStrategyRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type rolloutStrategyRouter struct {
	service rolloutStrategyServiceInterface
}

func newRolloutStrategyRouter(service rolloutStrategyServiceInterface) rolloutStrategyRouter {
	return rolloutStrategyRouter{
		service: service,
	}
}

// Implementation
func (r rolloutStrategyRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/use-cases/:useCaseId/rollout-strategy",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getRolloutStrategyInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getRolloutStrategyByID(ctx, request)
			if err == errRolloutStrategyNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "rollout-strategy-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.PUT(
		"/use-cases/:useCaseId/rollout-strategy",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request updateRolloutStrategyInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.updateRolloutStrategy(ctx, request)
			if err == errRolloutStrategyNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errRolloutStrategyWrongConfigFormat {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "rollout-strategy-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
