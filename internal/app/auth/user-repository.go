package auth

import (
	"gorm.io/gorm"
)

type authUserRepositoryInterface interface {
	findAuthUserByUsernameAndPassword(tx *gorm.DB, username string, password string) (authUserEntity, error)
	findAuthUserByUsername(tx *gorm.DB, username string) (authUserEntity, error)
}

type authUserRepository struct {
	readOnlyAuthUser  authUserEntity
	readWriteAuthUser authUserEntity
}

func newAuthUserRepository(readOnlyAuthUser authUserEntity, readWriteAuthUser authUserEntity) authUserRepository {
	return authUserRepository{
		readOnlyAuthUser:  readOnlyAuthUser,
		readWriteAuthUser: readWriteAuthUser,
	}
}

func (r authUserRepository) findAuthUserByUsernameAndPassword(tx *gorm.DB, username string, password string) (authUserEntity, error) {
	if username == r.readOnlyAuthUser.Username && password == r.readOnlyAuthUser.Password {
		return r.readOnlyAuthUser, nil
	} else if username == r.readWriteAuthUser.Username && password == r.readWriteAuthUser.Password {
		return r.readWriteAuthUser, nil
	}
	return authUserEntity{}, nil
}

func (r authUserRepository) findAuthUserByUsername(tx *gorm.DB, username string) (authUserEntity, error) {
	switch username {
	case r.readOnlyAuthUser.Username:
		return r.readOnlyAuthUser, nil
	case r.readWriteAuthUser.Username:
		return r.readWriteAuthUser, nil
	}
	return authUserEntity{}, nil
}
