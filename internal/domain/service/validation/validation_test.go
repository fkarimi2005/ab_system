package validation

import (
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
	"errors"
	"testing"
)

func variants(ws ...int) []models.ExperimentVariant {
	res := make([]models.ExperimentVariant, 0, len(ws))
	for i, w := range ws {
		res = append(res, models.ExperimentVariant{
			Name: string(rune('a' + i)), Value: "v", Weight: w, IsControl: i == 0,
		})
	}
	return res
}

func TestValidateConfig(t *testing.T) {
	noControl := variants(5000, 5000)
	noControl[0].IsControl = false
	twoControls := variants(5000, 5000)
	twoControls[1].IsControl = true
	dup := variants(5000, 5000)
	dup[1].Name = dup[0].Name

	tests := []struct {
		name     string
		expName  string
		audience int
		vs       []models.ExperimentVariant
		want     error
	}{
		{"ok", "exp", 10000, variants(5000, 5000), nil},
		{"ok partial audience", "exp", 2000, variants(1000, 1000), nil},
		{"empty name", "", 10000, variants(10000), errs.ErrExperimentNameIsEmpty},
		{"audience zero", "exp", 0, variants(10000), errs.ErrInvalidField},
		{"audience over 100%", "exp", 10001, variants(10001), errs.ErrInvalidField},
		{"no variants", "exp", 100, nil, errs.ErrExperimentVariantsEmpty},
		{"sum mismatch", "exp", 10000, variants(5000, 4000), errs.ErrWeightsSumMismatch},
		{"no control", "exp", 10000, noControl, errs.ErrControlVariantCount},
		{"two controls", "exp", 10000, twoControls, errs.ErrControlVariantCount},
		{"duplicate names", "exp", 10000, dup, errs.ErrInvalidField},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateConfig(tt.expName, tt.audience, tt.vs)
			if !errors.Is(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		typ, val string
		ok       bool
	}{
		{"bool", "true", true}, {"bool", "yes", false},
		{"number", "1.5", true}, {"number", "abc", false},
		{"string", "x", true}, {"json", "x", false},
	}
	for _, tt := range tests {
		if err := ValidateValue(tt.typ, tt.val); (err == nil) != tt.ok {
			t.Errorf("%s/%s: err=%v, ok=%v", tt.typ, tt.val, err, tt.ok)
		}
	}
}
