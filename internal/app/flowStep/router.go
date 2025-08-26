package flowStep

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type flowStepRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type flowStepRouter struct {
	service flowStepServiceInterface
}

func newFlowStepRouter(service flowStepServiceInterface) flowStepRouter {
	return flowStepRouter{
		service: service,
	}
}

// Implementation
func (r flowStepRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/flow-steps",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request ListFlowStepsInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			items, totalCount, err := r.service.listFlowSteps(ctx, request)
			if err == errFlowNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-step-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"items": items, "totalCount": totalCount, "hasNext": mm_router.HasNext(request.Page, request.PageSize, totalCount)})
		})

	router.GET(
		"/flow-steps/:flowStepID",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getFlowStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getFlowStepByID(ctx, request)
			if err == errFlowStepNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-step-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.PUT(
		"/flow-steps/:flowStepID",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request updateFlowStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.updateFlowStep(ctx, request)
			if err == errFlowStepNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errFlowStepWrongConfigFormat {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-step-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
