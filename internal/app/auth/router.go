package auth

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmrouter"
	"github.com/ai-model-match/backend/internal/pkg/mmtimeout"
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
		mmtimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input
			var request loginUserInputDto
			mmrouter.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mmrouter.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.login(ctx, request.Username, request.Password)
			if err == errInvalidCredentials {
				mmrouter.ReturnBadRequestError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mmrouter.ReturnGenericError(ctx)
				return
			}
			mmrouter.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/auth/refresh",
		mmtimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			var request refreshTokenInputDto
			mmrouter.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mmrouter.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.refreshToken(ctx, request.RefreshToken)
			if err == errExpiredRefreshToken {
				mmrouter.ReturnUnauthorizedError(ctx)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mmrouter.ReturnGenericError(ctx)
				return
			}
			mmrouter.ReturnOk(ctx, &gin.H{"item": item})
		})

	router.POST(
		"/auth/logout",
		mmtimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input
			var request refreshTokenInputDto
			mmrouter.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mmrouter.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			err := r.service.revokeRefreshToken(ctx, request.RefreshToken)
			if err == errExpiredRefreshToken {
				mmrouter.ReturnUnauthorizedError(ctx)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "auth-router"), zap.Error(err))
				mmrouter.ReturnGenericError(ctx)
				return
			}
			mmrouter.ReturnNoContent(ctx)
		})
}
