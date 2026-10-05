package validation

import (
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
	"strconv"
)

func ValidateVariantValues(valueType string, variants []models.ExperimentVariant) error {
	for _, v := range variants {
		if err := ValidateValue(valueType, v.Value); err != nil {
			return err
		}
	}
	return nil
}

func ValidateValue(valueType, value string) error {
	switch valueType {
	case "bool":
		if value != "true" && value != "false" {
			return errs.ErrInvalidField
		}
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return errs.ErrInvalidField
		}
	case "string":
		// любая непустая строка
	default:
		return errs.ErrInvalidField
	}
	return nil
}
