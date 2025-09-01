package picker

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type pickerRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type pickerRouter struct {
	service pickerServiceInterface
}

func newPickerRouter(service pickerServiceInterface) pickerRouter {
	return pickerRouter{
		service: service,
	}
}

// Implementation
func (r pickerRouter) register(router *gin.RouterGroup) {
	router.POST(
		"/picker",
		mm_auth.AuthMiddleware([]string{mm_auth.M2M_WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request pickerInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.pick(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errUseCaseNotAcive {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errUseCaseStepNotFound {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errFlowNotFound {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errFlowsNotAvailable {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errFallbackFlowNotAvailable {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "picker-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
