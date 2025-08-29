package auth

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type authRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type authRouter struct {
	service authServiceInterface
}

func newAuthRouter(service authServiceInterface) authRouter {
	return authRouter{
		service: service,
	}
}

// Implementation
func (r authRouter) register(router *gin.RouterGroup) {
	router.POST(
		"/auth/login",
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input
			var request loginUserInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.login(ctx, request.Username, request.Password)
			if err == errInvalidCredentials {
				mm_router.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/auth/refresh",
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			var request refreshTokenInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.refreshToken(ctx, request.RefreshToken)
			if err == errExpiredRefreshToken {
				mm_router.ReturnUnauthorizedError(ctx)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/auth/logout",
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input
			var request refreshTokenInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			err := r.service.revokeRefreshToken(ctx, request.RefreshToken)
			if err == errExpiredRefreshToken {
				mm_router.ReturnUnauthorizedError(ctx)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnNoContent(ctx)
		})
}
