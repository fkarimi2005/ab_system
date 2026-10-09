package service

import (
	"AB_system/internal/domain/decision"
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
	"context"
	"errors"

	"github.com/google/uuid"
)

// Источник значения в ответе Decide.
const (
	SourceDefault     = "default"      // эксперимента нет или пользователь не подошёл
	SourceExperiment  = "experiment"   // значение из варианта эксперимента
	SourceUnknownFlag = "unknown_flag" // флага с таким ключом нет
)

// Узкие интерфейсы: сервису нужны только эти методы репозиториев.
type flagByKeyFinder interface {
	GetFeatureFlagByKey(ctx context.Context, key string) (models.FeatureFlag, error)
}

type runningExperimentFinder interface {
	GetRunningExperimentByFlagID(ctx context.Context, flagID uuid.UUID) (models.Experiment, error)
}

type DecideService struct {
	flags       flagByKeyFinder
	experiments runningExperimentFinder
}

func NewDecideService(flags flagByKeyFinder, experiments runningExperimentFinder) *DecideService {
	return &DecideService{flags: flags, experiments: experiments}
}

type DecideInput struct {
	SubjectID  string
	Attributes map[string]string
	FlagKeys   []string
}

type DecideItem struct {
	FlagKey      string
	Value        *string // nil для unknown_flag
	ValueType    string
	Source       string
	DecisionID   *uuid.UUID // только для source == experiment
	ExperimentID *uuid.UUID
	Variant      string
}

// Decide возвращает решение по каждому запрошенному флагу, в порядке запроса.
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

	if !matchesTargeting(exp, in.Attributes) {
		return def, nil
	}

	bucket := decision.Bucket(exp.ID, in.SubjectID)
	variant, ok := decision.Select(exp.Variants, exp.AudienceBP, bucket)
	if !ok {
		return def, nil // вне аудитории эксперимента
	}

	decisionID := decision.ID(exp.ID, exp.Version, in.SubjectID, key)
	expID := exp.ID
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

// matchesTargeting — место для правил таргетинга (шаг 2 плана).
// Пока правил нет, эксперимент подходит всем.
func matchesTargeting(_ models.Experiment, _ map[string]string) bool {
	return true
}
