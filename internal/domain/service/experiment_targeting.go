package service

import (
	"AB_system/internal/domain/targeting"
	"AB_system/pkg/errs"
	"encoding/json"
	"fmt"
)

// normalizeTargeting проверяет правило и возвращает то, что нужно сохранить:
// nil для пустого правила (подходят все), иначе исходный JSON.
func normalizeTargeting(raw []byte) (json.RawMessage, error) {
	rule, err := targeting.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidTargeting, err)
	}
	if rule == nil {
		return nil, nil
	}
	return json.RawMessage(raw), nil
}
