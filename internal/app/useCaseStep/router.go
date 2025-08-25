package useCaseStep

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type useCaseStepRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type useCaseStepRouter struct {
	service useCaseStepServiceInterface
}

func newUseCaseStepRouter(service useCaseStepServiceInterface) useCaseStepRouter {
	return useCaseStepRouter{
		service: service,
	}
}

// Implementation
func (r useCaseStepRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/use-cases/:useCaseID/steps",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request ListUseCaseStepsInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			items, totalCount, err := r.service.listUseCaseSteps(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCaseStep-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"items": items, "totalCount": totalCount, "hasNext": mm_router.HasNext(request.Page, request.PageSize, totalCount)})
		})

	router.GET(
		"/use-cases/:useCaseID/steps/:useCaseStepID",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getUseCaseStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getUseCaseStepByID(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseStepNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCaseStep-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/use-cases/:useCaseID/steps",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request createUseCaseStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.createUseCaseStep(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseStepSameCodeAlreadyExists {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCaseStep-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.PUT(
		"/use-cases/:useCaseID/steps/:useCaseStepID",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request updateUseCaseStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.updateUseCaseStep(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseStepNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseStepSameCodeAlreadyExists {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCaseStep-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.DELETE(
		"/use-cases/:useCaseID/steps/:useCaseStepID",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request deleteUseCaseStepInputDto
			mm_router.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			_, err := r.service.deleteUseCaseStep(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseStepNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCaseStep-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnNoContent(ctx)
		})
}
