package auth

import (
	"github.com/ai-model-match/backend/internal/pkg/mmenv"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Init the module by registering new APIs
*/
func Init(envs *mmenv.Envs, dbStorage *gorm.DB, routerGroup *gin.RouterGroup) {
	zap.L().Info("Initialize Auth package...")
	var repository authRepositoryInterface
	var userRepository authUserRepositoryInterface
	var util authUtilInterface
	var service authServiceInterface
	var router authRouterInterface

	roUser := authUserEntity{
		Username: envs.AuthUserReadOnlyUsername,
		Password: envs.AuthUserReadOnlyPassword,
		Claims:   []Claim{READ},
	}

	rwUser := authUserEntity{
		Username: envs.AuthUserReadWriteUsername,
		Password: envs.AuthUserReadWritePassword,
		Claims:   []Claim{READ, WRITE},
	}

	repository = newAuthRepository()
	userRepository = newAuthUserRepository(roUser, rwUser)
	util = newAuthUtil(envs.AuthJwtSecret, envs.AuthJwtAccessTokenDuration, envs.AuthJwtRefreshTokenDuration)
	service = newAuthService(dbStorage, repository, userRepository, util)

	router = newAuthRouter(service)
	router.register(routerGroup)
	zap.L().Info("Auth package initialized")
}
