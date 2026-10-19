package service

import (
	"AB_system/internal/domain/decision"
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/internal/domain/targeting"
	Mid "AB_system/internal/http/middlewares"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
)

const (
	SourceDefault     = "default"
	SourceExperiment  = "experiment"
	SourceUnknownFlag = "unknown_flag"
)

type flagByKeyFinder interface {
	GetFeatureFlagByKey(ctx context.Context, key string) (models.FeatureFlag, error)
}

type runningExperimentFinder interface {
	GetRunningExperimentByFlagID(ctx context.Context, flagID uuid.UUID) (models.Experiment, error)
}

type DecideService struct {
	decisionRepo repository.ExperimentDecisionRepository
	flags        flagByKeyFinder
	experiments  runningExperimentFinder
	expRepo      repository.ExperimentVariantRepository
}

func NewDecideService(decisionRepo repository.ExperimentDecisionRepository, flags flagByKeyFinder, experiments runningExperimentFinder) *DecideService {
	return &DecideService{decisionRepo: decisionRepo, flags: flags, experiments: experiments}
}

type DecideInput struct {
	SubjectID  string
	Attributes map[string]string
	FlagKeys   []string
}

type DecideItem struct {
	FlagKey      string
	Value        *string
	ValueType    string
	Source       string
	DecisionID   *uuid.UUID
	ExperimentID *uuid.UUID
	Variant      string
}

func (s *DecideService) Decide(ctx context.Context, in DecideInput) ([]DecideItem, error) {
	items := make([]DecideItem, 0, len(in.FlagKeys))
	for _, key := range in.FlagKeys {
		item, err := s.decideOne(ctx, in, key)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *DecideService) decideOne(ctx context.Context, in DecideInput, key string) (DecideItem, error) {
	flag, err := s.flags.GetFeatureFlagByKey(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return DecideItem{FlagKey: key, Source: SourceUnknownFlag}, nil
		}
		return DecideItem{}, err
	}

	def := DecideItem{FlagKey: key, Value: &flag.DefaultValue, ValueType: flag.ValueType, Source: SourceDefault}

	exp, err := s.experiments.GetRunningExperimentByFlagID(ctx, flag.ID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return def, nil // активного эксперимента нет
		}
		return DecideItem{}, err
	}

	if !matchesTargeting(ctx, exp, in.Attributes) {
		return def, nil
	}

	bucket := decision.Bucket(exp.ID, in.SubjectID)
	var variant models.ExperimentVariant
	variant, ok := decision.Select(exp.Variants, exp.AudienceBP, bucket)
	if !ok {
		return def, nil // вне аудитории эксперимента
	}

	Mid.RecordExperimentRequest(exp.ID, variant.ID)
	decisionID := decision.ID(exp.ID, exp.Version, in.SubjectID, key)
	expID := exp.ID

	userID, err := uuid.Parse(in.SubjectID)
	if err != nil {
		return DecideItem{}, err
	}

	dec, err := s.decisionRepo.GetByID(ctx, decisionID)
	if err != nil {
		if !errors.Is(err, errs.ErrRecordNotFound) {
			return DecideItem{}, err
		}
		if err := s.decisionRepo.Save(ctx, models.ExperimentDecision{
			ID:           decisionID,
			ExperimentID: exp.ID,
			UserID:       userID,
			VariantID:    variant.ID,
		}); err != nil {
			if errors.Is(err, errs.ErrDuplicateEntry) {
				d, err := s.decisionRepo.GetByID(ctx, decisionID)
				if err != nil {
					return DecideItem{}, err
				}
				variant, err = s.expRepo.GetVariant(ctx, d.VariantID)
				if err != nil {
					return DecideItem{}, err
				}
			} else {
				return DecideItem{}, err
			}
		}
	} else {
		variant, err = s.expRepo.GetVariant(ctx, dec.VariantID)
		if err != nil {
			return DecideItem{}, err
		}
	}
	if exp.RevertedToControl {
		c, ok := controlVariant(exp.Variants)
		if !ok {
			return def, nil
		}
		variant = c
	}

	value := variant.Value

	return DecideItem{
		FlagKey:      key,
		Value:        &value,
		ValueType:    flag.ValueType,
		Source:       SourceExperiment,
		DecisionID:   &decisionID,
		ExperimentID: &expID,
		Variant:      variant.Name,
	}, nil
}

func matchesTargeting(ctx context.Context, exp models.Experiment, attrs map[string]string) bool {
	rule, err := targeting.Parse(exp.Targeting)
	if err != nil {
		slog.WarnContext(ctx, "invalid stored targeting rule", "experiment_id", exp.ID, "err", err)
		return false
	}
	return rule.Matches(attrs)
}
func controlVariant(vs []models.ExperimentVariant) (models.ExperimentVariant, bool) {
	for _, v := range vs {
		if v.IsControl {
			return v, true
		}
	}
	return models.ExperimentVariant{}, false
}
