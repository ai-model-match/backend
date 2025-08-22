package user

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmauth"
	"github.com/ai-model-match/backend/internal/pkg/mmrouter"
	"github.com/ai-model-match/backend/internal/pkg/mmtimeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type userRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type userRouter struct {
	service userServiceInterface
}

func newUserRouter(service userServiceInterface) userRouter {
	return userRouter{
		service: service,
	}
}

// Implementation
func (r userRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/users/:userID",
		mmauth.AuthMiddleware([]string{mmauth.READ}),
		mmtimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request getUserInputDto
			mmrouter.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				mmrouter.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getUserByID(ctx, request)
			if err == errUserNotFound {
				mmrouter.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "user-router"), zap.Error(err))
				mmrouter.ReturnGenericError(ctx)
				return
			}
			mmrouter.ReturnOk(ctx, &gin.H{"item": item})
		})
}
