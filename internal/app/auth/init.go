package auth

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_auth"
	"github.com/ai-model-match/backend/internal/pkg/mm_env"
	"github.com/ai-model-match/backend/internal/pkg/mm_scheduler"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs
*/
func Init(envs *mm_env.Envs, dbStorage *gorm.DB, cron *mm_scheduler.Scheduler, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize Auth package...")
	var repository authRepositoryInterface
	var userRepository authUserRepositoryInterface
	var util authUtilInterface
	var service authServiceInterface
	var scheduler authSchedulerInterface
	var router authRouterInterface

	// Define the available user roles
	roUser := authUserEntity{
		Username:    envs.AuthUserReadOnlyUsername,
		Password:    envs.AuthUserReadOnlyPassword,
		Permissions: []string{mm_auth.READ},
	}
	rwUser := authUserEntity{
		Username:    envs.AuthUserReadWriteUsername,
		Password:    envs.AuthUserReadWritePassword,
		Permissions: []string{mm_auth.READ, mm_auth.WRITE},
	}

	repository = newAuthRepository()
	userRepository = newAuthUserRepository(roUser, rwUser)
	util = newAuthUtil(envs.AuthJwtSecret, envs.AuthJwtAccessTokenDuration, envs.AuthJwtRefreshTokenDuration)
	service = newAuthService(dbStorage, repository, userRepository, util)
	scheduler = newAuthScheduler(dbStorage, cron, repository)
	scheduler.init()

	router = newAuthRouter(service)
	router.register(routerGroup)
	zap.L().Info("Auth package initialized")
}
