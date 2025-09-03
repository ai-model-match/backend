package feedback

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/ai-model-match/backend/internal/pkg/mm_pubsub"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs and PubSub consumers.
*/
func Init(envs *mm_env.Envs, dbStorage *gorm.DB, pubSubAgent *mm_pubsub.PubSubAgent, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize Feedback package...")
	var repository feedbackRepositoryInterface
	var service feedbackServiceInterface
	var router feedbackRouterInterface

	repository = newFeedbackRepository()
	service = newFeedbackService(dbStorage, pubSubAgent, repository)
	router = newFeedbackRouter(service)
	router.register(routerGroup)
	zap.L().Info("Feedback package initialized")
}
