package user

import (
	"time"

	"github.com/ai-model-match/backend/internal/pkg/mmdb"
	"github.com/google/uuid"
)

type userModel struct {
	id        uuid.UUID  `gorm:"primaryKey;column:id;type:varchar(36)"`
	email     string     `gorm:"column:email;type:varchar(255)"`
	firstname string     `gorm:"column:firstname;type:varchar(255)"`
	lastname  string     `gorm:"column:lastname;type:varchar(255)"`
	createdAt time.Time  `gorm:"column:created_at;type:timestamp;autoCreateTime:false"`
	updatedAt time.Time  `gorm:"column:updated_at;type:timestamp;autoUpdateTime:false"`
	deletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;autoDeleteTime:false"`
	createdBy uuid.UUID  `gorm:"column:created_by;type:varchar(36)"`
	updatedBy uuid.UUID  `gorm:"column:updated_by;type:varchar(36)"`
	deletedBy *uuid.UUID `gorm:"column:deleted_by;type:varchar(36)"`
}

func (m userModel) tableName() string {
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
