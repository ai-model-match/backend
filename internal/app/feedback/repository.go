package feedback

import (
	"github.com/ai-model-match/backend/internal/pkg/mm_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type feedbackRepositoryInterface interface {
	getPickerCorrelationByID(tx *gorm.DB, correlationId uuid.UUID) (pickerCorrelationEntity, error)
	saveFeedback(tx *gorm.DB, feedback feedbackEntity, operation mm_db.SaveOperation) (feedbackEntity, error)
}

type feedbackRepository struct {
}

func newFeedbackRepository() feedbackRepository {
	return feedbackRepository{}
}

func (r feedbackRepository) getPickerCorrelationByID(tx *gorm.DB, correlationID uuid.UUID) (pickerCorrelationEntity, error) {
	var model *pickerCorrelationModel
	query := tx.Where("id = ?", correlationID)
	result := query.Limit(1).Find(&model)
	if result.Error != nil {
		return pickerCorrelationEntity{}, result.Error
	}
	if result.RowsAffected == 0 {
		return pickerCorrelationEntity{}, nil
	}
	return model.toEntity(), nil
}

func (r feedbackRepository) saveFeedback(tx *gorm.DB, feedback feedbackEntity, operation mm_db.SaveOperation) (feedbackEntity, error) {
	var model = feedbackModel(feedback)
	var err error
	switch operation {
	case mm_db.Create:
		err = tx.Create(model).Error
	case mm_db.Update:
		err = tx.Updates(model).Error
	case mm_db.Upsert:
		err = tx.Save(model).Error
	}
	if err != nil {
		return feedbackEntity{}, err
	}
	return feedback, nil
}
