package user

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmdb"
	"github.com/google/uuid"
)

type userModel struct {
	ID        uuid.UUID  `gorm:"primaryKey;column:id;type:varchar(36)"`
	Email     string     `gorm:"column:email;type:varchar(255)"`
	Firstname string     `gorm:"column:firstname;type:varchar(255)"`
	Lastname  string     `gorm:"column:lastname;type:varchar(255)"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;autoDeleteTime:false"`
	CreatedBy uuid.UUID  `gorm:"column:created_by;type:varchar(36)"`
	UpdatedBy uuid.UUID  `gorm:"column:updated_by;type:varchar(36)"`
	DeletedBy *uuid.UUID `gorm:"column:deleted_by;type:varchar(36)"`
}

func (m userModel) TableName() string {
	return "mm_user"
}

func (m userModel) toEntity() userEntity {
	return userEntity(m)
}

type userOrderBy string

const (
	userOrderByFirstname userOrderBy = "firstname"
	userOrderByLastname  userOrderBy = "lastname"
	userOrderByEmail     userOrderBy = "email"
	userOrderByRelevance userOrderBy = mmdb.RelevanceField
)

var availableUserOrderBy = []interface{}{
	userOrderByFirstname,
	userOrderByLastname,
	userOrderByEmail,
	userOrderByRelevance,
}
