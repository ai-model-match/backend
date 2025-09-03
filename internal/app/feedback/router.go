package feedback

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_router"
	"github.com/ai-model-match/backend/internal/pkg/mm_timeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type feedbackRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type feedbackRouter struct {
	service feedbackServiceInterface
}

func newFeedbackRouter(service feedbackServiceInterface) feedbackRouter {
	return feedbackRouter{
		service: service,
	}
}

// Implementation
func (r feedbackRouter) register(router *gin.RouterGroup) {

	router.POST(
		"/feedbacks",
		mm_auth.AuthMiddleware([]string{mm_auth.M2M_READ, mm_auth.M2M_WRITE}),
		mm_timeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		func(ctx *gin.Context) {
			// Input validation
			var request createFeedbackInputDto
			if err := mm_router.BindParameters(ctx, &request); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			if err := request.validate(); err != nil {
				mm_router.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.createFeedback(ctx, request)
			if err == errCorrelationNotFound {
				mm_router.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "feedback-router"), zap.Error(err))
				mm_router.ReturnGenericError(ctx)
				return
			}
			mm_router.ReturnOk(ctx, &gin.H{"item": item})
		})

}
