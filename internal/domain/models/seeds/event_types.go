package seeds

import (
	"AB_system/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedEventTypes(db *gorm.DB) error {
	eTypes := []models.EventType{
		{
			Key:              "added_to_favorites",
			Name:             "Added to Favorites",
			RequiresExposure: true,
			IsExposure:       false,
		},
		{
			Key:              "click",
			Name:             "clicked button",
			RequiresExposure: true,
			IsExposure:       false,
		},
		{Key: "shown",
			Name:             "Show ",
			RequiresExposure: false,
			IsExposure:       true,
		},
		{
			Key:              "purchase",
			Name:             "Purchase ",
			RequiresExposure: true,
			IsExposure:       false,
		}, {

			Key:              "error",
			Name:             "Error ",
			RequiresExposure: false,
			IsExposure:       false,
		},
		{
			Key:              "api_request",
			Name:             "API Request",
			RequiresExposure: false,
			IsExposure:       false,
		},
	}
	for _, e := range eTypes {
		if err := db.Where("key = ?", e.Key).
			Assign(e).
			FirstOrCreate(&e).Error; err != nil {
			return err
		}
	}
	return nil
}
func GetEventTypesIDByKey(
	db *gorm.DB,
	key string,
) (uuid.UUID, error) {
	var eventType models.EventType

	err := db.
		Select("id").
		Where("key = ?", key).
		First(&eventType).Error

	if err != nil {
		return uuid.Nil, err
	}

	return eventType.ID, nil
}
