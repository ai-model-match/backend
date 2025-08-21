package user

import (
	"github.com/ai-model-match/backend/internal/pkg/mmenv"
	"github.com/ai-model-match/backend/internal/pkg/mmpubsub"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs and PubSub consumers.
*/
func Init(envs *mmenv.Envs, dbStorage *gorm.DB, pubSubAgent *mmpubsub.PubSubAgent, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize User package...")
	var repository userRepositoryInterface
	var service userServiceInterface
	var router userRouterInterface
	var consumer userConsumerInterface

	repository = newUserRepository(envs.SearchRelevanceThreshold)
	service = newUserService(dbStorage, pubSubAgent, repository)
	router = newUserRouter(service)
	consumer = newUserConsumer(pubSubAgent, service)
	consumer.subscribe()
	router.register(routerGroup)
	zap.L().Info("User package initialized")
}
