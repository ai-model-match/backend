package picker

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
	zap.L().Info("Initialize Picker package...")
	var repository pickerRepositoryInterface
	var service pickerServiceInterface
	var router pickerRouterInterface

	repository = newPickerRepository()
	service = newPickerService(dbStorage, pubSubAgent, repository)
	router = newPickerRouter(service)
	router.register(routerGroup)
	zap.L().Info("Picker package initialized")
}
