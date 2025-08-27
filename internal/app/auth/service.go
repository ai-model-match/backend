package auth

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/ai-model-match/backend/internal/pkg/mm_utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type authServiceInterface interface {
	login(ctx *gin.Context, username string, password string) (authTokenEntity, error)
	refreshToken(ctx *gin.Context, refreshToken string) (authTokenEntity, error)
	revokeRefreshToken(ctx *gin.Context, refreshToken string) error
}

type authService struct {
	storage        *gorm.DB
	repository     authRepositoryInterface
	userRepository authUserRepositoryInterface
	util           authUtilInterface
}

func newAuthService(storage *gorm.DB, repository authRepositoryInterface, userRepository authUserRepositoryInterface, util authUtilInterface) authService {
	return authService{
		storage:        storage,
		repository:     repository,
		userRepository: userRepository,
		util:           util,
	}
}

func (s authService) login(ctx *gin.Context, username string, password string) (authTokenEntity, error) {
	// Check if the user exists
	user, err := s.userRepository.findAuthUserByUsernameAndPassword(s.storage, username, password)
	if err != nil {
		return authTokenEntity{}, err
	}
	if mm_utils.IsEmpty(user) {
		return authTokenEntity{}, errInvalidCredentials
	}
	// Generate Access and Refresh Token
	token, err := s.util.generateToken(user)
	if err != nil {
		return authTokenEntity{}, err
	}
	// Store the Refresh token in DB for further requests
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		authEntity := authSessionEntity{
			ID:           token.RefreshTokenID,
			Username:     user.Username,
			CreatedAt:    token.RefreshTokenCreatedAt,
			ExpiresAt:    token.RefreshTokenExpiresAt,
			RefreshToken: token.RefreshToken,
		}
		if _, err := s.repository.saveAuthSessionEntity(tx, authEntity, mm_db.Create); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return authTokenEntity{}, errTransaction
	}
	return token, nil
}

func (s authService) refreshToken(ctx *gin.Context, refreshToken string) (authTokenEntity, error) {
	var token authTokenEntity
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the auth entity by Refresh token
		authEntity, err := s.repository.getAuthSessionEntityByRefreshToken(tx, refreshToken, true)
		if err != nil {
			return err
		}
		if mm_utils.IsEmpty(authEntity) {
			return errExpiredRefreshToken
		}
		// Find the user and its information like claims
		user, err := s.userRepository.findAuthUserByUsername(tx, authEntity.Username)
		if err != nil {
			return errExpiredRefreshToken
		}
		// Generate a new Access and Refresh token
		if token, err = s.util.generateToken(user); err != nil {
			return err
		}
		// Replace the Refresh token in the DB for further request
		authEntity.CreatedAt = token.RefreshTokenCreatedAt
		authEntity.ExpiresAt = token.RefreshTokenExpiresAt
		authEntity.RefreshToken = token.RefreshToken
		if _, err := s.repository.saveAuthSessionEntity(tx, authEntity, mm_db.Update); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return authTokenEntity{}, errTransaction
	}
	return token, nil
}

func (s authService) revokeRefreshToken(ctx *gin.Context, refreshToken string) error {
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		// Find the auth entity by Refresh token
		authEntity, err := s.repository.getAuthSessionEntityByRefreshToken(tx, refreshToken, true)
		if err != nil {
			return err
		}
		// If not found, return without error
		if mm_utils.IsEmpty(authEntity) {
			return errExpiredRefreshToken
		}
		// If found, delete it
		if err := s.repository.deleteAuthSessionEntity(tx, authEntity); err != nil {
			return err
		}
		return nil
	})
	if errTransaction != nil {
		return errTransaction
	}
	return nil
}
