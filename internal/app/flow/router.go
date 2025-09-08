package flow

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type flowRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type flowRouter struct {
	service flowServiceInterface
}

func newFlowRouter(service flowServiceInterface) flowRouter {
	return flowRouter{
		service: service,
	}
}

// Implementation
func (r flowRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/flows",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request ListFlowsInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			items, totalCount, err := r.service.listFlows(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"items": items, "totalCount": totalCount, "hasNext": mm_router.HasNext(request.Page, request.PageSize, totalCount)})
		})

	router.GET(
		"/flows/:flowId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getFlowInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getFlowByID(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errFlowNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/flows",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request createFlowInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.createFlow(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.PUT(
		"/flows/:flowId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request updateFlowInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.updateFlow(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errFlowNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errFlowCannotBeDeactivatedIfLastActive {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.DELETE(
		"/flows/:flowId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request deleteFlowInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			_, err := r.service.deleteFlow(ctx, request)
			if err == errFlowNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errFlowCannotBeDeletedIfActive {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnNoContent(ctx)
		})

	router.POST(
		"/flows/:flowId/clone",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request cloneFlowInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.cloneFlow(ctx, request)
			if err == errFlowNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "flow-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})
}
