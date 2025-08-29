package useCase

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type useCaseRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type useCaseRouter struct {
	service useCaseServiceInterface
}

func newUseCaseRouter(service useCaseServiceInterface) useCaseRouter {
	return useCaseRouter{
		service: service,
	}
}

// Implementation
func (r useCaseRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/use-cases",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request ListUseCasesInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			items, totalCount, err := r.service.listUseCases(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCase-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"items": items, "totalCount": totalCount, "hasNext": mm_router.HasNext(request.Page, request.PageSize, totalCount)})
		})

	router.GET(
		"/use-cases/:useCaseId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getUseCaseInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getUseCaseByID(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCase-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/use-cases",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request createUseCaseInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.createUseCase(ctx, request)
			if err == errUseCaseSameCodeAlreadyExists {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCase-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.PUT(
		"/use-cases/:useCaseId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request updateUseCaseInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.updateUseCase(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseSameCodeAlreadyExists {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errUseCaseCodeChangeNotAllowedWhileActive {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			if err == errUseCaseCannotBeActivatedWithoutFallbackFlow {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCase-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.DELETE(
		"/use-cases/:useCaseId",
		mm_auth.AuthMiddleware([]string{mm_auth.READ, mm_auth.WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request deleteUseCaseInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			_, err := r.service.deleteUseCase(ctx, request)
			if err == errUseCaseNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			if err == errUseCaseCannotBeDeletedWhileActive {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "useCase-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnNoContent(ctx)
		})
}
