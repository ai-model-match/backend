package healthCheck

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmrouter"
	"github.com/ai-model-match/backend/internal/pkg/mmtimeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type healthCheckRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type healthCheckRouter struct {
	service healthCheckServiceInterface
}

func newHealthCheckRouter(service healthCheckServiceInterface) healthCheckRouter {
	return healthCheckRouter{
		service: service,
	}
}

// Implementation
func (r healthCheckRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/health-check",
		mmtimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Business Logic
			isOk, err := r.service.checkConnection(ctx)

			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "health-check-router"), zap.Error(err))
				mmrouter.ReturnGenericError(ctx)
				return
			}
			mmrouter.ReturnOk(ctx, &gin.H{"ok": isOk})
		})
}
