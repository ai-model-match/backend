package auth

import (
	"time"

	"github.com/google/uuid"
)

type authModel struct {
	ID           uuid.UUID `gorm:"primaryKey;column:id;type:varchar(36)"`
	Username     string    `gorm:"column:username;type::varchar(255)"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp"`
	ExpiresAt    time.Time `gorm:"column:expires_at;type:timestamp"`
	RefreshToken string    `gorm:"column:refresh_token;type:text"`
}

func (m authModel) TableName() string {
	return "mm_auth"
}

func (m authModel) toEntity() authEntity {
	return authEntity(m)
}
