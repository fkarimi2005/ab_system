package service

import (
	"AB_system/internal/domain/targeting"
	"AB_system/pkg/errs"
	"encoding/json"
	"fmt"
)

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
